package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// EmailService handles email operations using ZeptoMail API
type EmailService struct {
	apiKey      string
	fromAddress string
	client      *http.Client
}

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

// EmailRecipient represents a recipient with email address
type EmailRecipient struct {
	EmailAddress EmailAddress `json:"email_address"`
}

// EmailRequest represents the request payload for ZeptoMail API
type EmailRequest struct {
	From     EmailAddress     `json:"from"`
	To       []EmailRecipient `json:"to"`
	Subject  string           `json:"subject"`
	HTMLBody string           `json:"htmlbody"`
}

// EmailResponse represents the response from ZeptoMail API
type EmailResponse struct {
	Data []struct {
		Code    string `json:"code"`
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"data"`
	Message string `json:"message"`
}

// NewEmailService creates a new email service instance
func NewEmailService(apiKey, fromAddress string) *EmailService {
	return &EmailService{
		apiKey:      apiKey,
		fromAddress: fromAddress,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendEmail sends an email using ZeptoMail API
func (e *EmailService) SendEmail(to, toName, subject, htmlBody string) error {
	// If no API key is configured, log the email instead of sending
	if e.apiKey == "" {
		log.Printf("EMAIL WOULD BE SENT TO: %s (%s)", to, toName)
		log.Printf("SUBJECT: %s", subject)
		log.Printf("BODY: %s", htmlBody)
		return nil
	}

	emailReq := EmailRequest{
		From: EmailAddress{
			Address: e.fromAddress,
		},
		To: []EmailRecipient{
			{
				EmailAddress: EmailAddress{
					Address: to,
					Name:    toName,
				},
			},
		},
		Subject:  subject,
		HTMLBody: htmlBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.zeptomail.com/v1.1/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Zoho-enczapikey "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var emailResp EmailResponse
		if err := json.NewDecoder(resp.Body).Decode(&emailResp); err == nil {
			return fmt.Errorf("email API error (status %d): %s", resp.StatusCode, emailResp.Message)
		}
		return fmt.Errorf("email API error with status: %d", resp.StatusCode)
	}

	var emailResp EmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if there are any errors in the response data
	for _, data := range emailResp.Data {
		if data.Status != "success" {
			return fmt.Errorf("email send failed: %s", data.Message)
		}
	}

	return nil
}
