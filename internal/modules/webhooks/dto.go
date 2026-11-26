package webhooks

import (
	"encoding/json"
	"time"
)

// CreateEndpointInput is the input to POST /webhooks/endpoints.
type CreateEndpointInput struct {
	Name   string `json:"name"   validate:"required,min=2,max=120"`
	Slug   string `json:"slug"   validate:"required,min=3,max=60,lowercase"`
	Secret string `json:"secret" validate:"omitempty,min=16,max=200"`
}

// EndpointResponse is the outward-facing shape of an Endpoint.
type EndpointResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toEndpointResponse(e *Endpoint) EndpointResponse {
	return EndpointResponse{
		ID:        e.ID.String(),
		Name:      e.Name,
		Slug:      e.Slug,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func toEndpointResponses(endpoints []Endpoint) []EndpointResponse {
	out := make([]EndpointResponse, len(endpoints))
	for i, e := range endpoints {
		out[i] = toEndpointResponse(&e)
	}
	return out
}

// EventResponse is the outward-facing shape of a captured Event.
type EventResponse struct {
	ID         string          `json:"id"`
	EndpointID string          `json:"endpoint_id"`
	Method     string          `json:"method"`
	Headers    json.RawMessage `json:"headers"`
	Body       string          `json:"body"`
	SourceIP   string          `json:"source_ip"`
	ReceivedAt time.Time       `json:"received_at"`
}

func toEventResponse(e *Event) EventResponse {
	return EventResponse{
		ID:         e.ID.String(),
		EndpointID: e.EndpointID.String(),
		Method:     e.Method,
		Headers:    json.RawMessage(e.Headers),
		Body:       e.Body,
		SourceIP:   e.SourceIP,
		ReceivedAt: e.ReceivedAt,
	}
}

func toEventResponses(events []Event) []EventResponse {
	out := make([]EventResponse, len(events))
	for i, e := range events {
		out[i] = toEventResponse(&e)
	}
	return out
}
