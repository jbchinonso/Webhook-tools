// Package async runs fire-and-forget work safely. The core rule: a spawned
// goroutine must never be able to crash the process. Every spawn recovers
// panics, logs them, and returns.
//
// Use this ONLY for work whose result the caller does not need — sending
// an email, writing an audit row that isn't part of a transaction, firing a
// webhook. Anything the response depends on must stay synchronous.
package async

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"

	"gateway/webhook/internal/commons/logger"
)

type Runner struct {
	log *slog.Logger
	wg  sync.WaitGroup
}

func NewRunner(log *slog.Logger) *Runner {
	return &Runner{log: log}
}

func (r *Runner) Go(ctx context.Context, name string, fn func(context.Context)) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer func() {
			if rec := recover(); rec != nil {
				r.log.ErrorContext(ctx, "async_panic",
					slog.String("job", name),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
			}
		}()
		fn(ctx)
	}()
}

func (r *Runner) Wait() { r.wg.Wait() }

// Detach returns a context that carries over the request_id from reqCtx
// but is independent of the request's cancellation.
func Detach(reqCtx context.Context) context.Context {
	ctx := context.Background()
	if id, ok := reqCtx.Value(logger.RequestIDKey).(string); ok && id != "" {
		ctx = context.WithValue(ctx, logger.RequestIDKey, id)
	}
	return ctx
}
