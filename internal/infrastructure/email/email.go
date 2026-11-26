// Package email abstracts outbound email. The interface is deliberately tiny
// — we send one thing: a password reset link. Anything richer gets its own
// interface later.
package email

import (
	"context"
	"log/slog"
)

// Sender is the port. Implementations are adapters.
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Noop is the default in dev. It logs the email body to slog at Info level
// so password-reset URLs are visible in the terminal without an SMTP server.
type Noop struct {
	Log *slog.Logger
}

func (n *Noop) Send(ctx context.Context, to, subject, body string) error {
	n.Log.InfoContext(ctx, "email_noop",
		slog.String("to", to),
		slog.String("subject", subject),
		slog.String("body", body),
	)
	return nil
}