# Tutorial

This is a tutorial walking a user on how can use the min instances feature on Cloud Functions to mitigate Cold Starts.

Let’s take a deeper look at min instances with a real-world use case: recording, transforming and serving a podcast. Podcasts are super popular, when you record a podcast, you need to get the audio in the right format (mp3, wav), and then make the podcast accessible so that users can easily access, download and listen to it. The application takes a recorded podcast, transcribes it from wav to text, stores it in Cloud Storage, and then emails an end user with a link to the transcribed file.

Now let’s consider building your podcast transformation application with and without min instances. 

## Approach 1: Base case, without min instances


In this approach, we use Cloud Functions and Google Cloud Workflows to chain together  three individual cloud functions. The first function (transcribe), transcribes the podcast, the second function (store-transcription) consumes the result of the first function in the workflow and stores it in Cloud Storage , and the third function (send-email), is triggered by Cloud Storage when the transcribed result is stored to send an email to the user to inform them that the workflow is complete.



## Creating the Transcribe Function

This is the function which takes in an audio podcast and transcribes it into a text file.

```
PROJECT_ID=$(gcloud config get-value core/project)
```

```
TRANSCRIBE_SERVICE_ACCOUNT_EMAIL="transcribe-function@${PROJECT_ID}.iam.gserviceaccount.com"
```

```
gcloud iam service-accounts create transcribe-function
```

```
gcloud functions deploy transcribe \
  --allow-unauthenticated \
  --entry-point Transcribe \
  --runtime go113 \
  --trigger-http \
  --service-account ${TRANSCRIBE_SERVICE_ACCOUNT_EMAIL} \
  --source transcribe
```

### Testing the Transcribe Function

```
TRANSCRIBE_URL=$(gcloud functions describe transcribe \
  --format='value(httpsTrigger.url)')
```

```
curl -X POST ${TRANSCRIBE_URL} \
  -o podcast.txt \
  --data-binary @podcast.wav
```

> Results

```
cat podcast.txt
```

```
What's up YouTube? I'm Kelsey and welcome to my channel. Before we dive in please be sure to smash that like button and subscribe so you don't miss future videos.
```

## Create the Store Transcription Function

This is the funciton which stores the transcribed podcast obtained from the transcribe function in a file in cloud storage. Once the file is stored in cloud storage, an event is fired to invoke a function which sends an email to the user.

```
PROJECT_ID=$(gcloud config get-value project)
```

```
STORE_TRANSCRIPTION_SERVICE_ACCOUNT_EMAIL="store-transcription-function@${PROJECT_ID}.iam.gserviceaccount.com"
```

```
gcloud iam service-accounts create store-transcription-function
```

Create storage bucket to hold mp3 files:

```
TRANSCRIPTION_UPLOAD_BUCKET_NAME="${PROJECT_ID}-transcriptions"
```

```
gsutil mb gs://${TRANSCRIPTION_UPLOAD_BUCKET_NAME}
```

```
gsutil iam ch \
  serviceAccount:${STORE_TRANSCRIPTION_SERVICE_ACCOUNT_EMAIL}:objectAdmin \
  gs://${TRANSCRIPTION_UPLOAD_BUCKET_NAME}
```

Deploy the `store-transcription` function:

```
gcloud functions deploy store-transcription \
  --allow-unauthenticated \
  --entry-point StoreTranscription \
  --runtime go113 \
  --trigger-http \
  --service-account ${STORE_TRANSCRIPTION_SERVICE_ACCOUNT_EMAIL} \
  --set-env-vars="TRANSCRIPTION_UPLOAD_BUCKET_NAME=${TRANSCRIPTION_UPLOAD_BUCKET_NAME}" \
  --source store-transcription
```

### Test Transcription Uploads

```
gsutil ls gs://${TRANSCRIPTION_UPLOAD_BUCKET_NAME}
```

```
STORE_TRANSCRIPTION_URL=$(gcloud functions describe store-transcription \
  --format='value(httpsTrigger.url)')
```

```
curl -X POST ${STORE_TRANSCRIPTION_URL} \
  --data-binary @podcast.txt
```

```
gsutil ls gs://${TRANSCRIPTION_UPLOAD_BUCKET_NAME}
```

> Output

```
gs://hightowerlabs-transcriptions/podcast.txt
```

## Create the Send Email Function

This is a function which sends an email to a user notifying the user that the transcription of the podcast has been completed.

```
PROJECT_ID=$(gcloud config get-value project)
```

```
SEND_EMAIL_FUNCTION_SERVICE_ACCOUNT_EMAIL="sendemail-function@${PROJECT_ID}.iam.gserviceaccount.com"
```

```
gcloud iam service-accounts create sendemail-function
```

```
gcloud functions deploy send-email \
  --allow-unauthenticated \
  --entry-point SendEmail \
  --runtime go113 \
  --trigger-resource ${TRANSCRIPTION_UPLOAD_BUCKET_NAME} \
  --trigger-event google.storage.object.finalize \
  --service-account ${SEND_EMAIL_FUNCTION_SERVICE_ACCOUNT_EMAIL} \
  --source send-email
```

### Testing the Send Email Function

```
STORE_TRANSCRIPTION_URL=$(gcloud functions describe store-transcription \
  --format='value(httpsTrigger.url)')
```

```
curl -X POST ${STORE_TRANSCRIPTION_URL} \
  --data-binary @podcast.txt
```

```
gcloud functions logs read send-email
```

> Output

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      send-email  lhbh4r703djk  2021-07-14 17:17:32.665  Function execution took 3006 ms, finished with status: 'ok'
       send-email  lhbh4r703djk  2021-07-14 17:17:32.664  Email sent successfully
       send-email  lhbh4r703djk  2021-07-14 17:17:29.663  Sending email...
       send-email  lhbh4r703djk  2021-07-14 17:17:29.663  Processing send email request
D      send-email  lhbh4r703djk  2021-07-14 17:17:29.661  Function execution started
```

## Create a Workflow

```
gcloud workflows deploy transcribe \
  --source workflow.yaml
```

```
gcloud workflows run transcribe
```

```
gcloud functions logs read transcribe
```

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      transcribe  6k258ardszq7  2021-08-13 06:26:05.026  Function execution took 6737 ms, finished with status code: 200
D      transcribe  6k258ardszq7  2021-08-13 06:25:58.290  Function execution started
```

```
gcloud functions logs read store-transcription
```

```
LEVEL  NAME                 EXECUTION_ID  TIME_UTC                 LOG
D      store-transcription  kunzo4g724ui  2021-08-13 06:26:08.075  Function execution took 2383 ms, finished with status code: 200
D      store-transcription  kunzo4g724ui  2021-08-13 06:26:05.692  Function execution started
```

```
gcloud functions logs read send-email
```

> Output

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      send-email  e0sdnt52vlcf  2021-08-13 06:26:22.532  Function execution took 3013 ms, finished with status: 'ok'
       send-email  e0sdnt52vlcf  2021-08-13 06:26:22.529  Email sent successfully
       send-email  e0sdnt52vlcf  2021-08-13 06:26:19.528  Sending email...
```

## Approach 2: Setting Min Instance Configuration with your functions

In this approach, we follow all the same steps as in Approach 1, with an addition of a set of min instances for each of the functions in the given workflow.

Deploy the `transcribe` function with min instances:

```
gcloud beta functions deploy transcribe \
  --allow-unauthenticated \
  --entry-point Transcribe \
  --runtime go113 \
  --trigger-http \
  --service-account ${TRANSCRIBE_SERVICE_ACCOUNT_EMAIL} \
  --source transcribe \
  --min-instances 5
```

Deploy the `store-transcription` function with min instances:

```
gcloud beta functions deploy store-transcription \
  --allow-unauthenticated \
  --entry-point StoreTranscription \
  --runtime go113 \
  --trigger-http \
  --service-account ${STORE_TRANSCRIPTION_SERVICE_ACCOUNT_EMAIL} \
  --set-env-vars="TRANSCRIPTION_UPLOAD_BUCKET_NAME=${TRANSCRIPTION_UPLOAD_BUCKET_NAME}" \
  --source store-transcription \
  --min-instances 5
```

Deploy the `send-email` function with min instances:

```
gcloud beta functions deploy send-email \
  --allow-unauthenticated \
  --entry-point SendEmail \
  --runtime go113 \
  --trigger-resource ${TRANSCRIPTION_UPLOAD_BUCKET_NAME} \
  --trigger-event google.storage.object.finalize \
  --service-account ${SEND_EMAIL_FUNCTION_SERVICE_ACCOUNT_EMAIL} \
  --source send-email \
  --min-instances 5
```


## Re-run the Workflow with Min Instances

Run it the first time to warm the instance, this ensures any bootstrap logic in the init() method is completed once. The instance at this point is prewarmed, and does not need to have the bootstrapping/init logic to be run again.
```
gcloud workflows run transcribe
```

Run the workflow again a second time.
```
gcloud workflows run transcribe
```


```
gcloud functions logs read transcribe
```

> Output

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      transcribe  yu0magytxdyb  2021-08-13 06:26:17.843  Function execution took 5005 ms, finished with status code: 200
D      transcribe  yu0magytxdyb  2021-08-13 06:26:12.839  Function execution started
```

```
gcloud functions logs read store-transcription
```

> Output

```
LEVEL  NAME                      EXECUTION_ID  TIME_UTC                 LOG\
D      store-transcription  ci24ipm9d1c9  2021-08-13 06:26:18.345  Function execution took 397 ms, finished with status code: 200
D      store-transcription  ci24ipm9d1c9  2021-08-13 06:26:17.948  Function execution started
```

```
gcloud functions logs read send-email
```

> Output

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      send-email  f0gtxaktn5ce  2021-08-13 06:26:22.527  Function execution took 3009 ms, finished with status: 'ok'
       send-email  f0gtxaktn5ce  2021-08-13 06:26:22.526  Email sent successfully
       send-email  f0gtxaktn5ce  2021-08-13 06:26:19.525  Sending email...

```
