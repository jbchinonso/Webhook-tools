package users

import "time"

// Public is the outward-facing shape of a user. Never includes PasswordHash.
type Public struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Role            Role       `json:"role"`
	Status          Status     `json:"status"`
	FullName        string     `json:"full_name"`
	Phone           string     `json:"phone"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func toPublic(u *User) Public {
	return Public{
		ID:              u.ID.String(),
		Email:           u.Email,
		Role:            u.Role,
		Status:          u.Status,
		FullName:        u.FullName,
		Phone:           u.Phone,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}

// UpdateProfile is the input to PATCH /me.
type UpdateProfile struct {
	FullName *string `json:"full_name" validate:"omitempty,min=2,max=120"`
	Phone    *string `json:"phone"     validate:"omitempty,min=6,max=20"`
}