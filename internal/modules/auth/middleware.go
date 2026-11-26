package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/jwt"
	"gateway/webhook/internal/commons/response"
)

const (
	ctxUserID    = "user_id"
	ctxUserRole  = "user_role"
	ctxSessionID = "session_id"
)

// RequireAuth validates the access token and stashes the identity in the
// gin context. It does NOT check session revocation — the access token is
// short-lived by design, and refusing it before expiry would require a DB
// round trip per request.
func RequireAuth(issuer *jwt.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, apperr.NewService("auth", apperr.KindUnauthorized, "auth.missing_token", "authentication required", nil))
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")

		claims, err := issuer.Verify(raw)
		if err != nil {
			code := "auth.invalid_token"
			if err == jwt.ErrExpired {
				code = "auth.expired_token"
			}
			response.Fail(c, apperr.NewService("auth", apperr.KindUnauthorized, code, "invalid or expired token", err))
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUserRole, claims.Role)
		c.Set(ctxSessionID, claims.SessionID)
		c.Next()
	}
}

// RequireRole gates a route on one of the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role, ok := c.Get(ctxUserRole)
		if !ok {
			response.Fail(c, apperr.NewService("auth", apperr.KindUnauthorized, "auth.unauthenticated", "not authenticated", nil))
			return
		}
		if _, ok := allowed[role.(string)]; !ok {
			response.Fail(c, apperr.NewService("auth", apperr.KindForbidden, "auth.forbidden", "insufficient permissions", nil))
			return
		}
		c.Next()
	}
}

// CurrentUserID returns the user ID placed by RequireAuth.
func CurrentUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// CurrentUserRole returns the role placed by RequireAuth.
func CurrentUserRole(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserRole)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
