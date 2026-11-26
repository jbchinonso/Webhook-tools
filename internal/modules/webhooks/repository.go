// Package webhooks owns the webhook_endpoints and webhook_events tables:
// registering capture endpoints and persisting whatever gets sent to them.
//
// TODO(delivery): forwarding a captured Event to a destination, retrying
// failed forwards, and replaying a past Event are not implemented yet.
// That will land as its own module (webhook_destinations + delivery_attempts
// tables) once this capture-and-store skeleton is in place.
package webhooks

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperr "gateway/webhook/internal/commons/errors"
)

// Repository is the port. The service only sees this interface.
type Repository interface {
	CreateEndpoint(ctx context.Context, e *Endpoint) error
	ListEndpoints(ctx context.Context) ([]Endpoint, error)
	FindEndpointByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)
	FindEndpointBySlug(ctx context.Context, slug string) (*Endpoint, error)

	CreateEvent(ctx context.Context, e *Event) error
	ListEventsByEndpoint(ctx context.Context, endpointID uuid.UUID) ([]Event, error)
	FindEventByID(ctx context.Context, id uuid.UUID) (*Event, error)
}

type repo struct {
	db *gorm.DB
}

// NewRepository takes the primary GORM handle.
func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) CreateEndpoint(ctx context.Context, e *Endpoint) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return apperr.NewRepo("Insert", "webhook_endpoints", err)
	}
	return nil
}

func (r *repo) ListEndpoints(ctx context.Context) ([]Endpoint, error) {
	var endpoints []Endpoint
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&endpoints).Error; err != nil {
		return nil, apperr.NewRepo("Select", "webhook_endpoints", err)
	}
	return endpoints, nil
}

func (r *repo) FindEndpointByID(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	var e Endpoint
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "webhook_endpoints", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "webhook_endpoints", err)
	}
	return &e, nil
}

func (r *repo) FindEndpointBySlug(ctx context.Context, slug string) (*Endpoint, error) {
	var e Endpoint
	err := r.db.WithContext(ctx).Where("slug = ?", slug).Take(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "webhook_endpoints", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "webhook_endpoints", err)
	}
	return &e, nil
}

func (r *repo) CreateEvent(ctx context.Context, e *Event) error {
	if err := r.db.WithContext(ctx).Create(e).Error; err != nil {
		return apperr.NewRepo("Insert", "webhook_events", err)
	}
	return nil
}

func (r *repo) ListEventsByEndpoint(ctx context.Context, endpointID uuid.UUID) ([]Event, error) {
	var events []Event
	err := r.db.WithContext(ctx).
		Where("endpoint_id = ?", endpointID).
		Order("received_at DESC").
		Find(&events).Error
	if err != nil {
		return nil, apperr.NewRepo("Select", "webhook_events", err)
	}
	return events, nil
}

func (r *repo) FindEventByID(ctx context.Context, id uuid.UUID) (*Event, error) {
	var e Event
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&e).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NewRepo("Select", "webhook_events", err).WithConstraint("not_found")
		}
		return nil, apperr.NewRepo("Select", "webhook_events", err)
	}
	return &e, nil
}
