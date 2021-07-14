package sendemail

import (
	"fmt"
	"net/http"
	"time"
)

func SendEmail(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing send email request\n")
	if err := sendEmail(); err != nil {
		http.Error(w, "Unable to send email", http.StatusInternalServerError)
		fmt.Printf("Unable to send email: %v", err)
	}
}

func sendEmail() error {
	fmt.Printf("Sending email...\n")
	time.Sleep(3 * time.Second)
	fmt.Printf("Email sent successfully\n")

	return nil
}
