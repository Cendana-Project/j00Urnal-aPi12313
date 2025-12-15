package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type Sender interface {
	Send(to, subject, htmlBody string) error
	SendWithContext(ctx context.Context, to, subject, htmlBody string) error
}

type SMTPSender struct {
	cfg *Config
}

func NewSMTPSender(cfg *Config) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	return s.SendWithContext(context.Background(), to, subject, htmlBody)
}

func (s *SMTPSender) SendWithContext(ctx context.Context, to, subject, htmlBody string) error {
	if s.cfg == nil || !s.cfg.Enabled {
		return nil // no-op jika dimatikan
	}
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// Determine 'from' email and name
	fromRaw := s.cfg.FromEmail
	if fromRaw == "" {
		fromRaw = s.cfg.Username
	}

	// Parse valid email address for envelope and header
	var envelopeFrom string
	var headerFrom string

	parsedAddr, err := mail.ParseAddress(fromRaw)
	if err == nil {
		envelopeFrom = parsedAddr.Address
		// Jika config punya FromName eksplisit, pakai itu. Jika tidak, pakai nama dari parse (jika ada).
		displayName := s.cfg.FromName
		if displayName == "" {
			displayName = parsedAddr.Name
		}
		headerFrom = formatAddress(envelopeFrom, displayName)
	} else {
		// Fallback jika parse gagal (misal cuma email string biasa tanpa <>)
		envelopeFrom = fromRaw
		headerFrom = formatAddress(fromRaw, s.cfg.FromName)
	}

	// Compose headers
	headers := map[string]string{
		"From":         headerFrom,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}
	var sb strings.Builder
	for k, v := range headers {
		sb.WriteString(k)
		sb.WriteString(": ")
		sb.WriteString(v)
		sb.WriteString("\r\n")
	}
	sb.WriteString("\r\n")
	sb.WriteString(htmlBody)
	msg := []byte(sb.String())

	// Dial TCP
	dialer := &net.Dialer{Timeout: s.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Quit()

	// STARTTLS (Gmail port 587)
	if s.cfg.UseSTARTTLS {
		tlsCfg := &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsCfg); err != nil {
				return err
			}
		}
	}

	// Auth
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	// From / To / Data
	// NOTE: Use envelopeFrom (pure email) for SMTP commands
	if err := client.Mail(envelopeFrom); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	return w.Close()
}

func formatAddress(email, name string) string {
	if name == "" {
		return email
	}
	// Basic encode for special chars if needed
	return fmt.Sprintf("%s <%s>", name, email)
}
