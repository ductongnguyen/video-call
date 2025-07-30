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

type WsNotificationHandler struct {
	usecase     signaling.UseCase
	mu          sync.Mutex
	connections map[uuid.UUID]map[uuid.UUID]*NotificationClient
}

func NewWsNotificationHandler(usecase signaling.UseCase) *WsNotificationHandler {
	return &WsNotificationHandler{
		usecase:     usecase,
		connections: make(map[uuid.UUID]map[uuid.UUID]*NotificationClient),
	}
}

type WebSocketMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type CallEventData struct {
	CallID string `json:"callId"`
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type NotificationClient struct {
	conn *websocket.Conn
	send chan []byte // Hàng đợi gửi tin nhắn để đảm bảo ghi an toàn, tránh race condition.
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

	client := &NotificationClient{
		conn: conn,
		send: make(chan []byte, 256),
	}
	connectionID := uuid.New()
	h.register(userIDUUID, connectionID, client)
	log.Printf("Notification connection established for user %s", userIDUUID)

	go h.writePump(userIDUUID, client)
	h.readPump(userIDUUID, connectionID, client)
}

func (h *WsNotificationHandler) register(userID, connectionID uuid.UUID, client *NotificationClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// KIỂM TRA PHÒNG THỦ #1: Map ngoài cùng có bị nil không?
	if h.connections == nil {
		log.Println("FATAL: h.connections map is nil! This should not happen if NewWsNotificationHandler was used.")
		// Khởi tạo lại nó để tránh panic, nhưng đây là một dấu hiệu của lỗi nghiêm trọng ở nơi khác.
		h.connections = make(map[uuid.UUID]map[uuid.UUID]*NotificationClient)
	}

	userConnections, ok := h.connections[userID]
	if !ok {
		// Tạo map con mới
		userConnections = make(map[uuid.UUID]*NotificationClient)
		// Gán map con vào map ngoài
		h.connections[userID] = userConnections
	}

	// KIỂM TRA PHÒNG THỦ #2: Map con có bị nil không?
	if userConnections == nil {
		log.Printf("FATAL: userConnections map is nil for user %s! Race condition suspected.", userID)
		// Dòng này sẽ gây panic, nhưng log ở trên sẽ cho bạn biết vấn đề.
	}

	// Dòng gốc gây panic
	userConnections[connectionID] = client
}

func (h *WsNotificationHandler) unregister(userID, connectionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Kiểm tra xem map của user có tồn tại không
	if userConnections, ok := h.connections[userID]; ok {
		// Kiểm tra xem connection cụ thể này có tồn tại không
		if client, ok := userConnections[connectionID]; ok {
			close(client.send)
			delete(userConnections, connectionID)
			
			log.Printf("Notification connection [%s] cleaned up for user [%s]", connectionID, userID)
		}

		// Nếu user không còn kết nối nào khác, xóa luôn map của user đó để tiết kiệm bộ nhớ
		if len(userConnections) == 0 {
			delete(h.connections, userID)
			log.Printf("User [%s] has no more connections, removing from map.", userID)
		}
	}
}

func (h *WsNotificationHandler) readPump(userID uuid.UUID, connectionID uuid.UUID, client *NotificationClient) {
	defer func() {
		h.unregister(userID, connectionID)
		log.Printf("Notification connection closed for use read %s", userID)
		client.conn.Close()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	err := client.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Printf("Failed to set read deadline for user %s: %v", userID, err)
	}
	client.conn.SetPongHandler(func(string) error {
		return client.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		messageType, payload, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message from user %s: %v", userID, err)
			}
			break
		}

		if messageType != websocket.TextMessage {
			log.Printf("Received non-text message from user %s: %v", userID, err)
			continue
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("Error parsing JSON from user %s: %v", userID, err)
			continue
		}
		switch msg.Event {
		case "accept_call":
			h.handleAcceptCall(userID, msg.Data)
		case "decline_call":
			h.handleDeclineCall(userID, msg.Data)
		case "end_call":
			h.handleEndCall(userID, msg.Data)
		case "webrtc_offer", "webrtc_answer", "ice_candidate":
			h.handleWebRTCSignal(userID, msg.Event, msg.Data)
		default:
			log.Printf("Received unknown event '%s' from user %s", msg.Event, userID)
		}
	}
}

func (h *WsNotificationHandler) writePump(userID uuid.UUID, client *NotificationClient) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error for user %s: %v", userID, err)
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error for user %s: %v", userID, err)
				return
			}
		}
	}
}

func (h *WsNotificationHandler) handleAcceptCall(userID uuid.UUID, data json.RawMessage) {
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
	if call.CalleeID != userID {
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
		"data": map[string]string{
			"callId":    callData.CallID,
			"startTime": time.Now().Format(time.RFC3339Nano),
		},
	})

	if err := h.SendMessageToUser(call.InitiatedID, notificationPayload); err != nil {
		log.Printf("Failed to notify caller %s about call acceptance: %v", call.InitiatedID, err)
	}

	if err := h.SendMessageToUser(call.CalleeID, notificationPayload); err != nil {
		log.Printf("Failed to notify callee %s about call acceptance: %v", call.CalleeID, err)
	}
}

func (h *WsNotificationHandler) handleDeclineCall(userID uuid.UUID, data json.RawMessage) {
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
	if call.InitiatedID != userID {
		if err := h.SendMessageToUser(call.InitiatedID, notificationPayload); err != nil {
			log.Printf("Failed to notify caller %s about call rejection: %v", userID, err)
		}
	} else {
		if err := h.SendMessageToUser(call.CalleeID, notificationPayload); err != nil {
			log.Printf("Failed to notify callee %s about call rejection: %v", userID, err)
		}
	}
}

func (h *WsNotificationHandler) handleEndCall(userID uuid.UUID, data json.RawMessage) {
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

	err = h.usecase.UpdateCallStatus(context.Background(), call.ID, models.CallStatusActive, models.CallStatusEnded, nil, nil)
	if err != nil {
		log.Printf("Error updating call %s: %v", callData.CallID, err)
		return
	}

	notificationPayload, _ := json.Marshal(map[string]interface{}{
		"event": "call_ended",
		"data":  map[string]string{"callId": callData.CallID},
	})
	if call.InitiatedID != userID {
		if err := h.SendMessageToUser(call.InitiatedID, notificationPayload); err != nil {
			log.Printf("Failed to notify caller %s about call rejection: %v", userID, err)
		}
	} else {
		if err := h.SendMessageToUser(call.CalleeID, notificationPayload); err != nil {
			log.Printf("Failed to notify callee %s about call rejection: %v", userID, err)
		}
	}
}

func (h *WsNotificationHandler) SendMessageToUser(userID uuid.UUID, msg []byte) error {
	h.mu.Lock()
	userConnections, ok := h.connections[userID]
	h.mu.Unlock()
	log.Printf("userConnections: %v", userConnections)

	if !ok {
		return fmt.Errorf("user %s is not connected, no connections found", userID)
	}

	// Lặp qua tất cả các kết nối của user và gửi tin nhắn
	for connID, client := range userConnections {
		log.Printf("Sending message to user [%s] on connection [%s]", userID, connID)
		client.send <- msg
	}
	return nil
}

type WebRTCSignalData struct {
	TargetID string          `json:"targetId"`
	Payload  json.RawMessage `json:"payload"`
}

func (h *WsNotificationHandler) handleWebRTCSignal(senderID uuid.UUID, event string, data json.RawMessage) {
	var signalData WebRTCSignalData
	log.Println("event: ", event)
	if err := json.Unmarshal(data, &signalData); err != nil {
		log.Printf("Error parsing WebRTC signal data from %s: %v", senderID, err)
		return
	}

	targetID, err := uuid.Parse(signalData.TargetID)
	if err != nil {
		log.Printf("Invalid TargetID '%s' in WebRTC signal from %s", signalData.TargetID, senderID)
		return
	}

	payloadToSend, err := json.Marshal(map[string]interface{}{
		"event": event,
		"data": map[string]interface{}{
			"senderId": senderID.String(),
			"payload":  signalData.Payload,
		},
	})
	if err != nil {
		log.Printf("Error marshaling WebRTC signal data from %s: %v", senderID, err)
		return
	}

	if err := h.SendMessageToUser(targetID, payloadToSend); err != nil {
		log.Printf("Failed to forward WebRTC signal from %s to %s: %v", senderID, targetID, err)
	}
}
