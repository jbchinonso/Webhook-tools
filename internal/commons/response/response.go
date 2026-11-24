// Package response defines the single JSON envelope used by every HTTP
// endpoint in the API:
//
//	{ "data": ..., "error": ..., "meta": ... }
//
// Exactly one of `data` or `error` is populated on any given response.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperr "gateway/webhook/internal/commons/errors"
)

const requestIDKey = "request_id"

type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  *Meta      `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Meta struct {
	RequestID string `json:"request_id,omitempty"`

	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: metaFrom(c)})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Data: data, Meta: metaFrom(c)})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Fail(c *gin.Context, err error) {
	he := apperr.ToHandler(err)
	c.AbortWithStatusJSON(he.Status, Envelope{
		Error: &ErrorBody{
			Code:    he.Code,
			Message: he.Message,
			Details: safeDetails(he),
		},
		Meta: metaFrom(c),
	})
}

func FailWithDetails(c *gin.Context, err error, details map[string]any) {
	he := apperr.ToHandler(err)
	c.AbortWithStatusJSON(he.Status, Envelope{
		Error: &ErrorBody{
			Code:    he.Code,
			Message: he.Message,
			Details: details,
		},
		Meta: metaFrom(c),
	})
}

func safeDetails(he *apperr.HandlerError) map[string]any {
	if he.Kind != apperr.KindValidation {
		return nil
	}
	return he.Metadata
}

func metaFrom(c *gin.Context) *Meta {
	m := &Meta{}
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			m.RequestID = s
		}
	}
	return m
}

func OKWithMeta(c *gin.Context, data any, meta *Meta) {
	if meta == nil {
		meta = &Meta{}
	}
	if id, ok := c.Get(requestIDKey); ok {
		if s, ok := id.(string); ok {
			meta.RequestID = s
		}
	}
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: meta})
}
