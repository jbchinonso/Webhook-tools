package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gateway/webhook/internal/commons/logger"
)

const (
	RequestIDKey    = "request_id"
	RequestIDHeader = "X-Request-ID"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)

		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, id)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
