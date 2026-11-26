package auth

import (
	"github.com/gin-gonic/gin"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/response"
	"gateway/webhook/internal/commons/validation"
	"gateway/webhook/internal/modules/users"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(c *gin.Context) {
	var in RegisterInput
	if !bindAndValidate(c, &in) {
		return
	}
	pair, u, err := h.svc.Register(c.Request.Context(), in, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, gin.H{
		"tokens": pair,
		"user":   publicUser(u),
	})
}

func (h *Handler) Login(c *gin.Context) {
	var in LoginInput
	if !bindAndValidate(c, &in) {
		return
	}
	pair, u, err := h.svc.Login(c.Request.Context(), in, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{
		"tokens": pair,
		"user":   publicUser(u),
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var in RefreshInput
	if !bindAndValidate(c, &in) {
		return
	}
	pair, err := h.svc.Refresh(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"tokens": pair})
}

func (h *Handler) Logout(c *gin.Context) {
	var in LogoutInput
	if !bindAndValidate(c, &in) {
		return
	}
	if err := h.svc.Logout(c.Request.Context(), in); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var in ForgotPasswordInput
	if !bindAndValidate(c, &in) {
		return
	}
	if err := h.svc.RequestPasswordReset(c.Request.Context(), in); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var in ResetPasswordInput
	if !bindAndValidate(c, &in) {
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), in); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"status": "ok"})
}

// bindAndValidate is a small helper that reads the JSON body and runs
// struct-tag validation, writing a 422 on failure and returning false.
func bindAndValidate(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.FailWithDetails(c,
			apperr.NewService("auth", apperr.KindValidation, "validation.invalid_body", "invalid request body", err),
			map[string]any{"_": []string{"malformed JSON"}},
		)
		return false
	}
	if details := validation.Validate(target); details != nil {
		m := make(map[string]any, len(details))
		for k, v := range details {
			m[k] = v
		}
		response.FailWithDetails(c,
			apperr.NewService("auth", apperr.KindValidation, "validation.failed", "validation failed", nil),
			m,
		)
		return false
	}
	return true
}

func publicUser(u *users.User) PublicUser {
	return PublicUser{
		ID:       u.ID.String(),
		Email:    u.Email,
		Role:     string(u.Role),
		FullName: u.FullName,
	}
}
