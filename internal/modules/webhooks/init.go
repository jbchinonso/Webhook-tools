package webhooks

import (
	"github.com/gin-gonic/gin"

	"gateway/webhook/internal/database"
)

// Init wires both route surfaces for this module:
//   - management, versioned under /api/v1 (api)
//   - public capture, unversioned on the bare engine (router) — capture URLs
//     are handed to third-party systems and must stay stable regardless of
//     API version.
func Init(router *gin.Engine, api *gin.RouterGroup, db *database.DB) {
	repo := NewRepository(db.Primary)
	svc := NewService(repo)
	h := NewHandler(svc)

	webhooksGroup := api.Group("/webhooks")
	{
		webhooksGroup.POST("/endpoints", h.CreateEndpoint)
		webhooksGroup.GET("/endpoints", h.ListEndpoints)
		webhooksGroup.GET("/endpoints/:id/events", h.ListEndpointEvents)
		webhooksGroup.GET("/events/:id", h.GetEvent)
	}

	router.Any("/hooks/:slug", h.Capture)
}
