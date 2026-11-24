package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	apperr "gateway/webhook/internal/commons/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ErrNotFound is returned by Find when the user doesn't exist.
var ErrNotFound = errors.New("user not found")

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewService("users", apperr.KindNotFound, "users.not_found", "user not found", err)
		}
		return nil, apperr.NewService("users", apperr.KindInternal, "users.lookup_failed", "could not load user", err)
	}
	return u, nil
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfile) (*User, error) {
	u, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.FullName != nil {
		u.FullName = strings.TrimSpace(*in.FullName)
	}
	if in.Phone != nil {
		u.Phone = strings.TrimSpace(*in.Phone)
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, apperr.NewService("users", apperr.KindInternal, "users.update_failed", "could not update profile", err)
	}
	return u, nil
}

func isNotFound(err error) bool {
	var re *apperr.RepoError
	if errors.As(err, &re) {
		return re.Constraint == "not_found"
	}
	return false
}
