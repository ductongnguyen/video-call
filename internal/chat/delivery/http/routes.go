package http

import (
	"github.com/ductongnguyen/vivy-chat/internal/chat"
	"github.com/ductongnguyen/vivy-chat/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Map news routes
func MapRoutes(group *gin.RouterGroup, h chat.Handlers, mw *middleware.MiddlewareManager) {
	group.POST("/conversations", h.CreateConversation)
	group.GET("/conversations", h.GetConversations)
	group.GET("/conversations/:id", h.GetConversation)
	group.PUT("/conversations/:id", h.UpdateConversation)
	group.DELETE("/conversations/:id", h.DeleteConversation)

	// Message routes within a conversation
	group.GET("/conversations/:id/messages", h.GetMessages)
	group.POST("/conversations/:id/messages", h.SendMessage)
}
