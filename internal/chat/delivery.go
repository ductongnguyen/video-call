package chat

import (
	"github.com/gin-gonic/gin"
)

// Auth HTTP Handlers interface
type Handlers interface {
	CreateConversation(c *gin.Context)
	GetConversation(c *gin.Context)
	GetConversations(c *gin.Context)
	UpdateConversation(c *gin.Context)
	DeleteConversation(c *gin.Context)

	GetMessages(c *gin.Context)
	SendMessage(c *gin.Context)
}
