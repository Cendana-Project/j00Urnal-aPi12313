package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BrevoAPISender sends emails via Brevo Transactional API v3 (not SMTP)
// This avoids SMTP port blocking issues common in cloud environments
type BrevoAPISender struct {
	apiKey  string
	from    string
	timeout time.Duration
}

type brevoPayload struct {
	Sender      brevoEmailAddress   `json:"sender"`
	To          []brevoEmailAddress `json:"to"`
	Subject     string              `json:"subject"`
	HTMLContent string              `json:"htmlContent"`
}

type brevoEmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func NewBrevoAPISender(apiKey, fromEmail string, timeout time.Duration) *BrevoAPISender {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &BrevoAPISender{
		apiKey:  apiKey,
		from:    fromEmail,
		timeout: timeout,
	}
}

func (s *BrevoAPISender) Send(to, subject, htmlBody string) error {
	return s.SendWithContext(context.Background(), to, subject, htmlBody)
}

func (s *BrevoAPISender) SendWithContext(ctx context.Context, to, subject, htmlBody string) error {
	payload := brevoPayload{
		Sender: brevoEmailAddress{
			Email: s.from,
			Name:  "MedikaOne",
		},
		To: []brevoEmailAddress{
			{Email: to},
		},
		Subject:     subject,
		HTMLContent: htmlBody,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: s.timeout}

	fmt.Printf("[Brevo API] Sending email to %s via HTTPS (api.brevo.com:443)...\n", to)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Brevo API error (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Brevo API] ✅ Email sent successfully to %s (status: %d)\n", to, resp.StatusCode)
	return nil
}

