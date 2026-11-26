package auth

import (
	"time"

	"github.com/google/uuid"
)

// Session is the persistence model for a refresh-token session. One row per
// successful login. Rotating a refresh token revokes the old row and inserts
// a new one.
type Session struct {
	ID               uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	UserID           uuid.UUID  `gorm:"column:user_id;type:uuid"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash"`
	UserAgent        string     `gorm:"column:user_agent"`
	IPAddress        string     `gorm:"column:ip_address"`
	ExpiresAt        time.Time  `gorm:"column:expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (Session) TableName() string { return "sessions" }

// PasswordResetToken is the persistence model for password reset tokens.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey;type:uuid"`
	UserID    uuid.UUID  `gorm:"column:user_id;type:uuid"`
	CodeHash  string     `gorm:"column:code_hash"`
	Attempts  int        `gorm:"column:attempts"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }