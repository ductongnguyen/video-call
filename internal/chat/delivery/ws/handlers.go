// internal/chat/delivery/ws_handler.go
package delivery

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"video-call/internal/chat"
	"video-call/internal/models"
	"video-call/pkg/utils"
	usecase "video-call/internal/chat/usecase"
	ws "video-call/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	hub       *ws.RedisHub
	chatUc    chat.UseCase
	writer    usecase.MessageWriterIface
	validator *validator.Validate
}

func NewWsHandler(hub *ws.RedisHub, chatUC chat.UseCase, writer usecase.MessageWriterIface) *WsHandler {
	return &WsHandler{
		hub:       hub,
		chatUc:    chatUC,
		writer:    writer,
		validator: validator.New(),
	}
}

// ServeWs handles WebSocket requests for Gin.
func (h *WsHandler) ServeWs(c *gin.Context) {
	// Get user ID from JWT token context
	userID, err := utils.GetUserIDFromCtx(c.Request.Context())
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized: missing or invalid token")
		return
	}

	// Get conversation ID from query parameter
	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		c.String(http.StatusBadRequest, "Missing conversation_id")
		return
	}

	// Convert user ID to string
	userIDStr := strconv.FormatUint(userID, 10)

	// Check if user is a participant in the conversation
	isParticipant, err := h.chatUc.IsUserInConversation(c.Request.Context(), userIDStr, conversationID)
	if err != nil {
		log.Printf("Failed to check conversation participation: %v", err)
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}
	if !isParticipant {
		c.String(http.StatusForbidden, "Forbidden: not a participant in this conversation")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	onMessage := func(message []byte) {
		var req CreateMessageRequest
		if err := json.Unmarshal(message, &req); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			return
		}

		if err := h.validator.Struct(req); err != nil {
			log.Printf("Invalid message format: %v", err)
			return
		}

		msg := models.Message{
			SenderID:       userIDStr,
			ConversationID: conversationID,
			Content:        req.Content,
			MessageType:    req.MessageType,
			Metadata:       req.Metadata,
		}

		// Đẩy vào hàng đợi DB
		h.writer.Enqueue(msg)

		// Publish lên Redis
		topic := fmt.Sprintf("conversation_%s", conversationID)
		err := h.hub.Publish(topic, message)
		if err != nil {
			log.Printf("Failed to publish message to Redis: %v", err)
		}
	}

	client := ws.NewClient(h.hub, conn, userIDStr, onMessage)

	h.hub.Register(client)
	topic := fmt.Sprintf("conversation_%s", conversationID)
	h.hub.SubscribeTopic(topic)
	if client.Topics == nil {
		client.Topics = make(map[string]bool)
	}
	client.Topics[topic] = true

	go client.WritePump()
	go client.ReadPump()

}
