# Tutorial


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
D      transcribe  e0vp3y8gfs5j  2021-07-14 17:29:52.490  Function execution took 5019 ms, finished with status code: 200
D      transcribe  e0vp3y8gfs5j  2021-07-14 17:29:47.471  Function execution started
```

```
gcloud functions logs read store-transcription
```

```
LEVEL  NAME                 EXECUTION_ID  TIME_UTC                 LOG
D      store-transcription  a83bkkpsx4jr  2021-07-14 17:29:52.859  Function execution took 259 ms, finished with status code: 200
D      store-transcription  a83bkkpsx4jr  2021-07-14 17:29:52.600  Function execution started
```

```
gcloud functions logs read send-email
```

> Output

```
LEVEL  NAME        EXECUTION_ID  TIME_UTC                 LOG
D      send-email  lhbhfloz53ii  2021-07-14 17:29:57.696  Function execution took 3005 ms, finished with status: 'ok'
       send-email  lhbhfloz53ii  2021-07-14 17:29:57.696  Email sent successfully
       send-email  lhbhfloz53ii  2021-07-14 17:29:54.695  Sending email...
       send-email  lhbhfloz53ii  2021-07-14 17:29:54.695  Processing send email request
D      send-email  lhbhfloz53ii  2021-07-14 17:29:54.692  Function execution started
```

## Setting Min Instance Configuration with your functions

Deploy the `transcribe` function with min instances:

```
gcloud beta functions deploy transcribe \
  --allow-unauthenticated \
  --entry-point Transcribe \
  --runtime go113 \
  --trigger-http \
  --service-account ${TRANSCRIBE_SERVICE_ACCOUNT_EMAIL} \
  --source transcribe
  --min_instances 3
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
  --source store-transcription
  --min_instances 3
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
  --source send-email
  --min_instances 3
```
