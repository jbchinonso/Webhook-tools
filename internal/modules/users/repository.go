package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "gateway/webhook/internal/commons/errors"
)

// Repository is the port. Everything downstream of the service only sees
// this interface, so swapping GORM for something else is a single-file change.
type Repository interface {
	Create(ctx context.Context, tx *gorm.DB, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, u *User) error
}

type repo struct {
	db *gorm.DB
}

// NewRepository takes the primary GORM handle. Writes and read-your-writes
// both go to primary.
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(ctx context.Context, tx *gorm.DB, u *User) error {
	conn := r.db
	if tx != nil {
		conn = tx
	}
	if err := conn.WithContext(ctx).Create(u).Error; err != nil {
		return apperr.NewRepo("Insert", "users", err)
	}
	return nil
}

func (r *repo) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "users", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "users", err)
	}
	return &u, nil
}

func (r *repo) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", email).
		Take(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "users", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "users", err)
	}
	return &u, nil
}

func (r *repo) Update(ctx context.Context, u *User) error {
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		return apperr.NewRepo("Update", "users", err)
	}
	return nil
}
