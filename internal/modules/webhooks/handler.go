package webhooks

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperr "gateway/webhook/internal/commons/errors"
	"gateway/webhook/internal/commons/response"
	"gateway/webhook/internal/commons/validation"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) CreateEndpoint(c *gin.Context) {
	var in CreateEndpointInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithDetails(c,
			apperr.NewService("webhooks", apperr.KindValidation, "validation.invalid_body", "invalid request body", err),
			map[string]any{"_": []string{"malformed JSON"}},
		)
		return
	}
	if details := validation.Validate(&in); details != nil {
		response.FailWithDetails(c,
			apperr.NewService("webhooks", apperr.KindValidation, "validation.failed", "validation failed", nil),
			toAnyMap(details),
		)
		return
	}

	e, err := h.svc.CreateEndpoint(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, toEndpointResponse(e))
}

func (h *Handler) ListEndpoints(c *gin.Context) {
	endpoints, err := h.svc.ListEndpoints(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toEndpointResponses(endpoints))
}

func (h *Handler) ListEndpointEvents(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.NewService("webhooks", apperr.KindValidation, "webhooks.invalid_endpoint_id", "invalid endpoint id", err))
		return
	}

	if _, err := h.svc.GetEndpoint(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}

	events, err := h.svc.ListEvents(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toEventResponses(events))
}

func (h *Handler) GetEvent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, apperr.NewService("webhooks", apperr.KindValidation, "webhooks.invalid_event_id", "invalid event id", err))
		return
	}

	e, err := h.svc.GetEvent(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toEventResponse(e))
}

// Capture is the public inbound route third-party systems send webhooks to.
// It accepts any method and body and stores the delivery as-is.
func (h *Handler) Capture(c *gin.Context) {
	slug := c.Param("slug")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.Fail(c, apperr.NewService("webhooks", apperr.KindValidation, "webhooks.invalid_body", "could not read request body", err))
		return
	}

	_, err = h.svc.Capture(c.Request.Context(), slug, c.Request.Method, c.Request.Header, string(body), c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}

func toAnyMap(e validation.Errors) map[string]any {
	out := make(map[string]any, len(e))
	for k, v := range e {
		out[k] = v
	}
	return out
}
