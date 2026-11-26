package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/response"
)

// Recovery converts a panic into a structured 500 response and a log line.
// Panics are NEVER used for control flow in this codebase; this middleware
// exists only as a last-resort safety net.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(c.Request.Context(), "panic_recovered",
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				response.Fail(c, apperr.New(
					apperr.KindInternal, "internal.panic", "internal server error",
				))
			}
		}()
		c.Next()
	}
}
