package signaling

import (
	"github.com/gin-gonic/gin"
)

type Handlers interface {
	CreateOrJoinCall(c *gin.Context)
}