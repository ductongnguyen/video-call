package http

import (
	"video-call/internal/auth"
	"video-call/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Map news routes
func MapRoutes(group *gin.RouterGroup, h auth.Handlers, mw *middleware.MiddlewareManager) {
	group.POST("/register", h.Register)
	group.POST("/login", h.Login)
	group.POST("/refresh", h.RefreshToken)
	group.Use(mw.AuthJWTMiddleware())
	group.GET("/users", h.GetUsers)
	group.GET("/users/:userId", h.GetUserByID)
}
