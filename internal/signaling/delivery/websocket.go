package delivery

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
	usecase signaling.CallUsecase
	mu      sync.Mutex
	peers   map[uuid.UUID]map[string]*websocket.Conn // callID -> userID -> conn
}

func NewWsHandler(usecase signaling.CallUsecase) *WsHandler {
	return &WsHandler{
		usecase: usecase,
		peers:   make(map[uuid.UUID]map[string]*websocket.Conn),
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
		// Forward message to peer
		h.forwardToPeer(callID, userID, msg)
		// Nếu là leave thì break
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