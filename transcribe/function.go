package transcribe

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func init() {
	// Sleep during cold starts.
	time.Sleep(2 * time.Second)
}

func Transcribe(w http.ResponseWriter, r *http.Request) {
	data, err := wavToText(r.Body)
	if err != nil {
		http.Error(w, "Unable to parse wav file", http.StatusBadRequest)
		log.Printf("Error parsing wav file: %v", err)
		return
	}

	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain")
	w.Write(data)
}

func wavToText(r io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("Error parsing wav file: %v", err)
	}
	time.Sleep(5 * time.Second)

	return data, nil
}
