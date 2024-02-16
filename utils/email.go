package db

import (
  "bytes"
  "encoding/json"
  "fmt"
  "os"
  "io"
  "net/http"
)

func SendEmail(email string, code string) EmailResponse {
  apiURL := os.Getenv("EMAIL_API_URL")
  apiKey := os.Getenv("RAPID_API_KEY")
  replyTo := os.Getenv("REPLY_TO_EMAIL")
	var emailResp EmailResponse

	payload := bytes.NewBufferString(fmt.Sprintf(`{
		"sendto": "%s",
		"name": "Custom Name Here",
		"replyTo": "%s",
		"ishtml": false,
		"title": "BsonDB Verification Code",
		"body": "Your verification code is: %s"
	}`, email, replyTo, code))

	req, err := http.NewRequest("POST", apiURL, payload)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RapidAPI-Key", apiKey)
	req.Header.Set("X-RapidAPI-Host", "mail-sender-api1.p.rapidapi.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to send email: %v", err)}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return EmailResponse{Error: true, Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	err = json.Unmarshal(body, &emailResp)
	if err != nil {
    return EmailResponse{Error: true, Message: fmt.Sprintf("failed to unmarshal response: %v", err)}
	}

	return emailResp 
}
