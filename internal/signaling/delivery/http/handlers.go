package http

import (
	"net/http"
	"video-call/internal/signaling"
	"video-call/pkg/logger"
    "encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	signalingWs "video-call/internal/signaling/delivery/ws"
)

// Handler handles HTTP requests for chat features
type Handler struct {
	useCase signaling.UseCase
	wsNotificationHandler *signalingWs.WsNotificationHandler
	logger  logger.Logger
}

// NewHandler creates a new chat HTTP handler
func NewHandler(useCase signaling.UseCase, wsNotificationHandler *signalingWs.WsNotificationHandler, logger logger.Logger) *Handler {
	return &Handler{
		useCase: useCase,
		wsNotificationHandler: wsNotificationHandler,
		logger:  logger,
	}
}

func (h *Handler) CreateOrJoinCall(c *gin.Context) {
	var req callRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	callerID, err := uuid.Parse(req.CallerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid caller_id"})
		return
	}
	calleeID, err := uuid.Parse(req.CalleeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callee_id"})
		return
	}
	call, role, err := h.useCase.CreateOrJoinCall(c.Request.Context(), callerID, calleeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notificationPayload := gin.H{
		"type": "ringing",
		"data": gin.H{
			"call_id": call.ID.String(),
			"caller": gin.H{
				"id": call.CallerID.String(),
			},
		},
	}

	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal notification payload"})
		return
	}
	h.wsNotificationHandler.SendMessageToUser(calleeID, payloadBytes)

	c.JSON(http.StatusOK, callResponse{Call: call, Role: role})
}
