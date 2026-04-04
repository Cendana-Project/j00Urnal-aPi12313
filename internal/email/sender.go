package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

type Sender interface {
	Send(to, subject, htmlBody string) error
	SendWithContext(ctx context.Context, to, subject, htmlBody string) error
	// SendWithContextAndText sends HTML with an optional plain-text part (used by Brevo API; SMTP ignores text when empty).
	SendWithContextAndText(ctx context.Context, to, subject, htmlBody, textBody string) error
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

func (s *SMTPSender) SendWithContextAndText(ctx context.Context, to, subject, htmlBody, _ string) error {
	return s.SendWithContext(ctx, to, subject, htmlBody)
}

func (s *SMTPSender) SendWithContext(ctx context.Context, to, subject, htmlBody string) error {
	if s.cfg == nil || !s.cfg.Enabled {
		return nil // no-op jika dimatikan
	}

	// Debug logging
	fmt.Printf("[SMTP] Attempting to send email to %s via %s:%d (timeout: %v)\n",
		to, s.cfg.Host, s.cfg.Port, s.cfg.Timeout)

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

	// Compose headers (Subject: RFC 2047 when non-ASCII)
	subjectHeader := subject
	if subjectNeedsEncodedWord(subject) {
		subjectHeader = mime.QEncoding.Encode("utf-8", subject)
	}
	headers := map[string]string{
		"From":         headerFrom,
		"To":           to,
		"Subject":      subjectHeader,
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

	// Determine if we need SSL/TLS directly (port 465) or STARTTLS (port 587)
	useSSL := s.cfg.Port == 465

	var client *smtp.Client
	if useSSL {
		// Port 465: Use TLS connection directly (SSL)
		fmt.Printf("[SMTP] Using SSL/TLS (port 465)\n")
		tlsCfg := &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		fmt.Printf("[SMTP] Dialing TLS connection to %s...\n", addr)
		tlsConn, err := tls.DialWithDialer(&net.Dialer{Timeout: s.cfg.Timeout}, "tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer tlsConn.Close()

		client, err = smtp.NewClient(tlsConn, s.cfg.Host)
		if err != nil {
			return err
		}
		defer client.Quit()
	} else {
		// Port 587: Use STARTTLS
		fmt.Printf("[SMTP] Using STARTTLS (port 587)\n")
		dialer := &net.Dialer{Timeout: s.cfg.Timeout}
		fmt.Printf("[SMTP] Dialing TCP connection to %s...\n", addr)
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP dial failed: %w", err)
		}
		defer conn.Close()
		fmt.Printf("[SMTP] TCP connection established\n")

		client, err = smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
		defer client.Quit()
		fmt.Printf("[SMTP] SMTP client created\n")

		// STARTTLS (for port 587)
		if s.cfg.UseSTARTTLS {
			fmt.Printf("[SMTP] Initiating STARTTLS...\n")
			tlsCfg := &tls.Config{
				ServerName: s.cfg.Host,
				MinVersion: tls.VersionTLS12,
			}
			if ok, _ := client.Extension("STARTTLS"); ok {
				if err := client.StartTLS(tlsCfg); err != nil {
					return fmt.Errorf("STARTTLS failed: %w", err)
				}
				fmt.Printf("[SMTP] STARTTLS successful\n")
			} else {
				fmt.Printf("[SMTP] STARTTLS not supported by server\n")
			}
		}
	}

	// Auth
	fmt.Printf("[SMTP] Authenticating as %s...\n", s.cfg.Username)
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	if ok, _ := client.Extension("AUTH"); ok {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		fmt.Printf("[SMTP] Authentication successful\n")
	} else {
		fmt.Printf("[SMTP] No AUTH extension, skipping authentication\n")
	}

	// From / To / Data
	// NOTE: Use envelopeFrom (pure email) for SMTP commands
	fmt.Printf("[SMTP] Setting envelope FROM: %s\n", envelopeFrom)
	if err := client.Mail(envelopeFrom); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	fmt.Printf("[SMTP] Setting envelope TO: %s\n", to)
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	fmt.Printf("[SMTP] Sending email data...\n")
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return fmt.Errorf("write message failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close data writer failed: %w", err)
	}

	fmt.Printf("[SMTP] Email sent successfully to %s\n", to)
	return nil
}

func formatAddress(email, name string) string {
	if name == "" {
		return email
	}
	// Basic encode for special chars if needed
	return fmt.Sprintf("%s <%s>", name, email)
}

func subjectNeedsEncodedWord(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
		if r < 32 && r != '\t' {
			return true
		}
	}
	return false
}
