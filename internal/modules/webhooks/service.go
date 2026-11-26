package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"

	apperr "gateway/webhook/internal/commons/errors"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateEndpoint(ctx context.Context, in CreateEndpointInput) (*Endpoint, error) {
	e := &Endpoint{
		ID:     uuid.New(),
		Name:   strings.TrimSpace(in.Name),
		Slug:   strings.ToLower(strings.TrimSpace(in.Slug)),
		Secret: in.Secret,
		Status: EndpointStatusActive,
	}
	if err := s.repo.CreateEndpoint(ctx, e); err != nil {
		if isUniqueViolation(err) {
			return nil, apperr.NewService("webhooks", apperr.KindConflict, "webhooks.slug_taken", "slug already in use", err)
		}
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.create_endpoint_failed", "could not create endpoint", err)
	}
	return e, nil
}

func (s *Service) ListEndpoints(ctx context.Context) ([]Endpoint, error) {
	endpoints, err := s.repo.ListEndpoints(ctx)
	if err != nil {
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.list_endpoints_failed", "could not list endpoints", err)
	}
	return endpoints, nil
}

func (s *Service) GetEndpoint(ctx context.Context, id uuid.UUID) (*Endpoint, error) {
	e, err := s.repo.FindEndpointByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewService("webhooks", apperr.KindNotFound, "webhooks.endpoint_not_found", "endpoint not found", err)
		}
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.lookup_endpoint_failed", "could not load endpoint", err)
	}
	return e, nil
}

// Capture persists one inbound delivery against the endpoint identified by
// slug. It does not verify the endpoint's secret against a signature, and it
// does not forward the event anywhere — see the module-level TODO.
func (s *Service) Capture(ctx context.Context, slug, method string, headers map[string][]string, body, sourceIP string) (*Event, error) {
	endpoint, err := s.repo.FindEndpointBySlug(ctx, slug)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewService("webhooks", apperr.KindNotFound, "webhooks.endpoint_not_found", "endpoint not found", err)
		}
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.lookup_endpoint_failed", "could not load endpoint", err)
	}
	if endpoint.Status != EndpointStatusActive {
		return nil, apperr.NewService("webhooks", apperr.KindForbidden, "webhooks.endpoint_disabled", "endpoint is disabled", nil)
	}

	headerJSON, err := json.Marshal(headers)
	if err != nil {
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.encode_headers_failed", "could not encode headers", err)
	}

	e := &Event{
		ID:         uuid.New(),
		EndpointID: endpoint.ID,
		Method:     method,
		Headers:    headerJSON,
		Body:       body,
		SourceIP:   sourceIP,
	}
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.capture_failed", "could not capture event", err)
	}
	return e, nil
}

func (s *Service) ListEvents(ctx context.Context, endpointID uuid.UUID) ([]Event, error) {
	events, err := s.repo.ListEventsByEndpoint(ctx, endpointID)
	if err != nil {
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.list_events_failed", "could not list events", err)
	}
	return events, nil
}

func (s *Service) GetEvent(ctx context.Context, id uuid.UUID) (*Event, error) {
	e, err := s.repo.FindEventByID(ctx, id)
	if err != nil {
		if isNotFound(err) {
			return nil, apperr.NewService("webhooks", apperr.KindNotFound, "webhooks.event_not_found", "event not found", err)
		}
		return nil, apperr.NewService("webhooks", apperr.KindInternal, "webhooks.lookup_event_failed", "could not load event", err)
	}
	return e, nil
}

func isNotFound(err error) bool {
	var re *apperr.RepoError
	if errors.As(err, &re) {
		return re.Constraint == "not_found"
	}
	return false
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}
