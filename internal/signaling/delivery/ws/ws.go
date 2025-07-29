package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"video-call/internal/models"
	"video-call/internal/signaling"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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
		conn:    make(map[uuid.UUID]*websocket.Conn),
	}
}

type WebSocketMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type CallEventData struct {
	CallID string `json:"callId"`
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
		var m struct {
			Type string `json:"type"`
		}
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

	// defer func() {
	// 	h.mu.Lock()
	// 	if _, ok := h.conn[userIDUUID]; ok {
	// 		delete(h.conn, userIDUUID)
	// 	}
	// 	h.mu.Unlock()
	// 	conn.Close()
	// 	log.Printf("Notification connection cleaned up for user %s", userIDUUID)
	// }()

	h.mu.Lock()
	h.conn[userIDUUID] = conn
	h.mu.Unlock()
	log.Printf("Notification connection established for user %s", userIDUUID)

	const (
		// Thời gian chờ server ghi một tin nhắn đến client.
		writeWait = 10 * time.Second

		// Thời gian chờ client trả lời Pong. Phải lớn hơn pongWait.
		pongWait = 60 * time.Second

		// Tần suất gửi Ping đến client. Phải nhỏ hơn pongWait.
		pingPeriod = (pongWait * 9) / 10
	)

	conn.SetReadDeadline(time.Now().Add(pongWait))
	// Khi nhận được Pong, ta sẽ cập nhật lại deadline này.
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Goroutine để gửi Ping định kỳ
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for range ticker.C {
			// Đặt deadline để tránh bị block mãi mãi nếu kết nối chết.
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error for user %s: %v", userIDUUID, err)
				return // Thoát khỏi goroutine nếu có lỗi
			}
		}
	}()

	// Use a for range loop to read messages from the WebSocket connection
	for {
		messageType, payload, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message from user %s: %v", userIDUUID, err)
			}
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}
		var msg WebSocketMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("Error parsing JSON from user %s: %v", userIDUUID, err)
			continue
		}

		switch msg.Event {
		case "accept_call":
			h.handleAcceptCall(userIDUUID, msg.Data)
		case "decline_call":
			h.handleDeclineCall(userIDUUID, msg.Data)
		default:
			log.Printf("Received unknown event '%s' from user %s", msg.Event, userIDUUID)
		}
	}
}

func (h *WsNotificationHandler) handleAcceptCall(calleeID uuid.UUID, data json.RawMessage) {
	var callData CallEventData
	if err := json.Unmarshal(data, &callData); err != nil {
		log.Printf("Error parsing accept_call data: %v", err)
		return
	}

	callID := uuid.MustParse(callData.CallID)
	call, err := h.usecase.GetCallByID(context.Background(), callID)
	if err != nil {
		log.Printf("Error getting call %s: %v", callID, err)
		return
	}
	if call.CalleeID != calleeID {
		return
	}
	if call.Status != models.CallStatusRinging {
		return
	}

	err = h.usecase.UpdateCallStatus(context.Background(), call.ID, models.CallStatusRinging, models.CallStatusActive, nil, nil)
	if err != nil {
		log.Printf("Error updating call %s: %v", callData.CallID, err)
		return
	}

	notificationPayload, _ := json.Marshal(map[string]interface{}{
		"event": "call_accepted",
		"data":  map[string]string{"callId": callData.CallID, "calleeId": calleeID.String()},
	})

	callerIDFromDB := call.InitiatedID
	if err := h.SendMessageToUser(callerIDFromDB, notificationPayload); err != nil {
		log.Printf("Failed to notify caller %s about call acceptance: %v", callerIDFromDB, err)
	}
}

func (h *WsNotificationHandler) handleDeclineCall(calleeID uuid.UUID, data json.RawMessage) {
	var callData CallEventData
	if err := json.Unmarshal(data, &callData); err != nil {
		log.Printf("Error parsing decline_call data: %v", err)
		return
	}

	callID := uuid.MustParse(callData.CallID)
	call, err := h.usecase.GetCallByID(context.Background(), callID)
	if err != nil {
		log.Printf("Error getting call %s: %v", callData.CallID, err)
		return
	}

	err = h.usecase.UpdateCallStatus(context.Background(), call.ID, models.CallStatusRinging, models.CallStatusRejected, nil, nil)
	if err != nil {
		log.Printf("Error updating call %s: %v", callData.CallID, err)
		return
	}

	notificationPayload, _ := json.Marshal(map[string]interface{}{
		"event": "call_declined",
		"data":  map[string]string{"callId": callData.CallID},
	})

	if err := h.SendMessageToUser(call.InitiatedID, notificationPayload); err != nil {
		log.Printf("Failed to notify caller %s about call rejection: %v", call.CallerID, err)
	}
}

func (h *WsNotificationHandler) SendMessageToUser(userID uuid.UUID, msg []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.conn[userID]; ok {
		return conn.WriteMessage(websocket.TextMessage, msg)
	}

	log.Printf("User %s is not connected for notifications.", userID)
	fmt.Println(h.conn)
	return nil
}
