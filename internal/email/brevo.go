package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const brevoMaxErrBodyRunes = 600

// BrevoAPISender sends emails via Brevo Transactional API v3 (not SMTP)
// This avoids SMTP port blocking issues common in cloud environments
type BrevoAPISender struct {
	apiKey   string
	from     string
	fromName string
	client   *http.Client
}

type brevoPayload struct {
	Sender      brevoEmailAddress   `json:"sender"`
	To          []brevoEmailAddress `json:"to"`
	Subject     string              `json:"subject"`
	HTMLContent string              `json:"htmlContent"`
	TextContent string              `json:"textContent,omitempty"`
}

type brevoEmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func NewBrevoAPISender(apiKey, fromEmail, fromName string, timeout time.Duration) *BrevoAPISender {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	fromAddr, parsedName := normalizeBrevoFrom(fromEmail)
	if strings.TrimSpace(fromName) == "" && parsedName != "" {
		fromName = parsedName
	}
	return &BrevoAPISender{
		apiKey:   strings.TrimSpace(apiKey),
		from:     fromAddr,
		fromName: strings.TrimSpace(fromName),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// normalizeBrevoFrom trims, strips common .env quotes, and extracts the address via RFC 5322 parsing.
func normalizeBrevoFrom(raw string) (email string, displayName string) {
	s := strings.TrimSpace(raw)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if s == "" {
		return "", ""
	}
	if a, err := mail.ParseAddress(s); err == nil && strings.TrimSpace(a.Address) != "" {
		return strings.TrimSpace(a.Address), strings.TrimSpace(a.Name)
	}
	return s, ""
}

// brevoFromLooksPlausible rejects values that slip through config as placeholders or garbage.
func brevoFromLooksPlausible(addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return false
	}
	if strings.ContainsAny(addr, " \t\r\n<>") {
		return false
	}
	if _, err := mail.ParseAddress(addr); err != nil {
		return false
	}
	return strings.Contains(addr, "@")
}

func (s *BrevoAPISender) Send(to, subject, htmlBody string) error {
	return s.SendWithContext(context.Background(), to, subject, htmlBody)
}

func (s *BrevoAPISender) SendWithContext(ctx context.Context, to, subject, htmlBody string) error {
	return s.SendWithContextAndText(ctx, to, subject, htmlBody, "")
}

func (s *BrevoAPISender) SendWithContextAndText(ctx context.Context, to, subject, htmlBody, textBody string) error {
	if strings.TrimSpace(s.apiKey) == "" {
		return fmt.Errorf("brevo: API key is not configured")
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("brevo: recipient address is empty")
	}
	if strings.TrimSpace(s.from) == "" {
		return fmt.Errorf("brevo: sender address is not configured")
	}
	if !brevoFromLooksPlausible(s.from) {
		return fmt.Errorf("brevo: sender %q is not a valid From address; set SMTP_FROM_EMAIL to a verified sender in Brevo", s.from)
	}

	sender := brevoEmailAddress{Email: s.from}
	if s.fromName != "" {
		sender.Name = s.fromName
	}
	payload := brevoPayload{
		Sender: sender,
		To: []brevoEmailAddress{
			{Email: to},
		},
		Subject:     subject,
		HTMLContent: htmlBody,
	}
	if t := strings.TrimSpace(textBody); t != "" {
		payload.TextContent = t
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("brevo: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("brevo: create request: %w", err)
	}

	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	logrus.WithFields(logrus.Fields{
		"component": "brevo",
		"to":        maskRecipientForLog(to),
	}).Debug("sending transactional email")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("brevo: read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		bodyStr := truncateRunes(string(body), brevoMaxErrBodyRunes)
		if resp.StatusCode == 400 && strings.Contains(bodyStr, "valid sender email") {
			return fmt.Errorf("brevo: Brevo rejected sender %s (add/verify it under Senders, domains & dedicated IPs in Brevo; then set SMTP_FROM_EMAIL to that exact address): %s",
				maskRecipientForLog(s.from), bodyStr)
		}
		return fmt.Errorf("brevo: API error status=%d body=%s", resp.StatusCode, bodyStr)
	}

	logrus.WithFields(logrus.Fields{
		"component":   "brevo",
		"to":          maskRecipientForLog(to),
		"status_code": resp.StatusCode,
	}).Info("transactional email accepted by provider")
	return nil
}

func maskRecipientForLog(addr string) string {
	addr = strings.TrimSpace(addr)
	at := strings.LastIndex(addr, "@")
	if at <= 1 || at >= len(addr)-1 {
		return "***"
	}
	local, domain := addr[:at], addr[at+1:]
	if len(local) < 2 {
		return "***@" + domain
	}
	return local[:2] + "***@" + domain
}

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
