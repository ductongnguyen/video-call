package http

import (
	"video-call/internal/middleware"
	"video-call/internal/signaling"

	"github.com/gin-gonic/gin"
)

// Map news routes
func MapRoutes(group *gin.RouterGroup, h signaling.Handlers, mw *middleware.MiddlewareManager) {
	group.POST("/call", h.CreateOrJoinCall)

}
