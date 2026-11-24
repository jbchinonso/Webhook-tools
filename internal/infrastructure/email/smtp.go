package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTP is a minimal SMTPSender. It uses net/smtp, which is fine for a
// portfolio project but lacks STARTTLS negotiation subtleties for some
// providers. When you need production-grade SMTP, swap the adapter — the
// interface stays.
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func (s *SMTP) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	headers := []string{
		"From: " + s.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	var auth smtp.Auth
	if s.Username != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}
	return smtp.SendMail(addr, auth, s.From, []string{to}, []byte(msg))
}