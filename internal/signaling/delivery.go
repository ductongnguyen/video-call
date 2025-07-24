package signaling

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

type CallRESTDelivery interface {
    RegisterRoutes(rg *gin.RouterGroup)
}

type CallWebSocketDelivery interface {
    HandleWebSocket(conn *websocket.Conn)
} 