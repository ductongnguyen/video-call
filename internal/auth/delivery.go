package auth

import (
	"github.com/gin-gonic/gin"
)

// Auth HTTP Handlers interface
type Handlers interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	GetUserByID(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetUsers(c *gin.Context)
}
