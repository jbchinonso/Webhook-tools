package health

import (
	"github.com/gin-gonic/gin"

	"gateway/webhook/internal/commons/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Liveness(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

func (h *Handler) Readiness(c *gin.Context) {
	report, err := h.svc.Readiness(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, report)
}
