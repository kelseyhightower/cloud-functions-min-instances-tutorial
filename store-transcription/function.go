package storetranscription

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

func init() {
	bucketName = os.Getenv("TRANSCRIPTION_UPLOAD_BUCKET_NAME")
	if bucketName == "" {
		fmt.Printf("Bucket name is empty\n")
	}

	ctx := context.Background()

	var err error
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		fmt.Printf("Error initializing storage client: %v\n", err)
		return
	}

	time.Sleep(2 * time.Second)
}

var (
	storageClient *storage.Client
	bucketName    string
)

func StoreTranscription(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to parse transcription file", http.StatusBadRequest)
		fmt.Printf("Error parsing transcription file: %v\n", err)
		http.Error(w, "Error processing transcription file", 500)
		return
	}

	defer r.Body.Close()

	ctx := context.Background()
	wc := storageClient.Bucket(bucketName).Object("podcast.txt").NewWriter(ctx)
	wc.Write(data)
	if err := wc.Close(); err != nil {
		log.Printf("Error uploading transcription file to %s: %v", bucketName, err)
		http.Error(w, "Error processing transcription file", 500)
		return
	}
}
