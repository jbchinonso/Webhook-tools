package middleware

import (
	"github.com/gin-gonic/gin"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/response"
)

// RateLimitByIP returns a gin middleware that calls allow(key) for the
// client's IP and rejects the request when it returns false.
// A nil allow function disables the limiter (no-op middleware).
func RateLimitByIP(allow func(string) bool, code, message string) gin.HandlerFunc {
	if allow == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if !allow(c.ClientIP()) {
			response.Fail(c, apperr.NewService("rate_limit", apperr.KindUnavailable, code, message, nil))
			return
		}
		c.Next()
	}
}
