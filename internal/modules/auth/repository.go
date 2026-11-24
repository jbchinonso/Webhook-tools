package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "gateway/webhook/internal/commons/errors"
)

// Repository is the port for session and password-reset persistence.
// User lookup lives in the users module; the auth service depends on that
// module's Repository interface directly.
type Repository interface {
	// Sessions
	CreateSession(ctx context.Context, s *Session) error
	FindSessionByTokenHash(ctx context.Context, hash string) (*Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error

	// Password reset
	InvalidateActiveResetTokens(ctx context.Context, userID uuid.UUID) error
	CreateResetToken(ctx context.Context, t *PasswordResetToken) error
	FindActiveResetTokenByUser(ctx context.Context, userID uuid.UUID) (*PasswordResetToken, error)
	IncrementResetAttempts(ctx context.Context, id uuid.UUID) error
	MarkResetTokenUsed(ctx context.Context, id uuid.UUID) error
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

// --- Sessions ---

func (r *repo) CreateSession(ctx context.Context, s *Session) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return apperr.NewRepo("Insert", "sessions", err)
	}
	return nil
}

func (r *repo) FindSessionByTokenHash(ctx context.Context, hash string) (*Session, error) {
	var s Session
	err := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", hash).
		Take(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "sessions", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "sessions", err)
	}
	return &s, nil
}

func (r *repo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
	if err != nil {
		return apperr.NewRepo("Update", "sessions", err)
	}
	return nil
}

// --- Password reset ---

func (r *repo) CreateResetToken(ctx context.Context, t *PasswordResetToken) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return apperr.NewRepo("Insert", "password_reset_tokens", err)
	}
	return nil
}

func (r *repo) FindActiveResetTokenByUser(ctx context.Context, userID uuid.UUID) (*PasswordResetToken, error) {
	var t PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND used_at IS NULL", userID).
		Order("created_at DESC").
		Take(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "password_reset_tokens", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "password_reset_tokens", err)
	}
	return &t, nil
}

func (r *repo) MarkResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now).Error
	if err != nil {
		return apperr.NewRepo("Update", "password_reset_tokens", err)
	}
	return nil
}

func (r *repo) InvalidateActiveResetTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
	if err != nil {
		return apperr.NewRepo("Update", "password_reset_tokens", err)
	}
	return nil
}

func (r *repo) IncrementResetAttempts(ctx context.Context, id uuid.UUID) error {
	err := r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + 1")).Error
	if err != nil {
		return apperr.NewRepo("Update", "password_reset_tokens", err)
	}
	return nil
}
