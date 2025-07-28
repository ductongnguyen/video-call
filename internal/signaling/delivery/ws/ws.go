package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"video-call/internal/signaling"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type WsHandler struct {
	usecase signaling.UseCase
	mu      sync.Mutex
	peers   map[uuid.UUID]map[string]*websocket.Conn // callID -> userID -> conn
}

func NewWsHandler(usecase signaling.UseCase) *WsHandler {
	return &WsHandler{
		usecase: usecase,
		peers:   make(map[uuid.UUID]map[string]*websocket.Conn),
	}
}

type WsNotificationHandler struct {
	usecase signaling.UseCase
	mu      sync.Mutex
	conn    map[uuid.UUID]*websocket.Conn // callID -> conn
}

func NewWsNotificationHandler(usecase signaling.UseCase) *WsNotificationHandler {
	return &WsNotificationHandler{
		usecase: usecase,
		conn:   make(map[uuid.UUID]*websocket.Conn),
	}
}

func (h *WsHandler) ServeWs(c *gin.Context) {
	userID := c.Query("user_id")
	callIDStr := c.Query("call_id")
	if userID == "" || callIDStr == "" {
		c.String(http.StatusBadRequest, "Missing user_id or call_id")
		return
	}
	callID, err := uuid.Parse(callIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid call_id")
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	if h.peers[callID] == nil {
		h.peers[callID] = make(map[string]*websocket.Conn)
	}
	h.peers[callID][userID] = conn
	h.mu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		h.forwardToPeer(callID, userID, msg)
		var m struct{ Type string `json:"type"` }
		_ = json.Unmarshal(msg, &m)
		if m.Type == "leave" {
			break
		}
	}

	h.mu.Lock()
	delete(h.peers[callID], userID)
	if len(h.peers[callID]) == 0 {
		delete(h.peers, callID)
	}
	h.mu.Unlock()
}

func (h *WsHandler) forwardToPeer(callID uuid.UUID, fromUserID string, msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	peers := h.peers[callID]
	for userID, peerConn := range peers {
		if userID != fromUserID {
			_ = peerConn.WriteMessage(websocket.TextMessage, msg)
		}
	}
} 

func (h *WsNotificationHandler) ServeWsNotifications(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.String(http.StatusBadRequest, "Missing user_id")
		return
	}
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid user_id")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade notifications connection for user %s: %v", userID, err)
		return
	}

	defer func() {
		h.mu.Lock()
		delete(h.conn,userIDUUID)
		h.mu.Unlock()
		conn.Close() 
		log.Printf("Notification connection cleaned up for user %s", userIDUUID)
	}()

	h.mu.Lock()
	h.conn[userIDUUID] = conn
	h.mu.Unlock()
	log.Printf("Notification connection established for user %s", userIDUUID)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

}

func (h *WsNotificationHandler) SendMessageToUser(userID uuid.UUID, msg []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Tìm kết nối của người dùng trong map.
	if conn, ok := h.conn[userID]; ok {
		// Nếu tìm thấy, gửi tin nhắn đi.
		return conn.WriteMessage(websocket.TextMessage, msg)
	}

	// Nếu không tìm thấy, trả về lỗi (người dùng không online).
	log.Printf("User %s is not connected for notifications.", userID)
	return nil // Hoặc trả về một lỗi cụ thể nếu cần
}