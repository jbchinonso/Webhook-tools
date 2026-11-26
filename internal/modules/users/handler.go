package users

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/response"
	"gateway/webhook/internal/commons/validation"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Me returns the authenticated user's profile.
//
// The user ID is placed in the gin context by the auth middleware.
func (h *Handler) Me(c *gin.Context) {
	id, ok := userIDFromContext(c)
	if !ok {
		response.Fail(c, apperr.NewService("users", apperr.KindUnauthorized, "auth.unauthenticated", "not authenticated", nil))
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toPublic(u))
}

// UpdateMe patches the profile.
func (h *Handler) UpdateMe(c *gin.Context) {
	id, ok := userIDFromContext(c)
	if !ok {
		response.Fail(c, apperr.NewService("users", apperr.KindUnauthorized, "auth.unauthenticated", "not authenticated", nil))
		return
	}

	var in UpdateProfile
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithDetails(c,
			apperr.NewService("users", apperr.KindValidation, "validation.invalid_body", "invalid request body", err),
			map[string]any{"_": []string{"malformed JSON"}},
		)
		return
	}
	if details := validation.Validate(&in); details != nil {
		response.FailWithDetails(c,
			apperr.NewService("users", apperr.KindValidation, "validation.failed", "validation failed", nil),
			toAnyMap(details),
		)
		return
	}

	u, err := h.svc.UpdateProfile(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toPublic(u))
}

// userIDFromContext reads the user ID that RequireAuth placed in the context.
// Defined here so users.Handler stays independent of the auth package.
func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, false
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func toAnyMap(e validation.Errors) map[string]any {
	out := make(map[string]any, len(e))
	for k, v := range e {
		out[k] = v
	}
	return out
}
