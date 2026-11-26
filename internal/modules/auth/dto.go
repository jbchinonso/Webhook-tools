package auth

import "time"

// RegisterInput is the request body for POST /auth/register.
// Role is deliberately constrained to attendee|organizer — admin is never
// self-assignable.
type RegisterInput struct {
	Email    string `json:"email"     validate:"required,email,max=254"`
	Password string `json:"password"  validate:"required,password"`
	FullName string `json:"full_name" validate:"required,min=2,max=120"`
	Phone    string `json:"phone"     validate:"omitempty,min=6,max=20"`
	Role     string `json:"role"      validate:"required,oneof=attendee organizer"`
}

// LoginInput is the request body for POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshInput is the request body for POST /auth/refresh.
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutInput is the request body for POST /auth/logout.
type LogoutInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ForgotPasswordInput is the request body for POST /auth/password/forgot.
type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordInput is the request body for POST /auth/password/reset.
type ResetPasswordInput struct {
	Email       string `json:"email"        validate:"required,email"`
	Code        string `json:"code"         validate:"required,len=6,numeric"`
	NewPassword string `json:"new_password" validate:"required,password"`
}

// TokenPair is what we return from register/login/refresh. The access token
// lives in the response body; clients store the refresh token in secure
// storage or an HttpOnly cookie (their responsibility, not ours).
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// PublicUser is the subset of user data returned with tokens.
type PublicUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
}