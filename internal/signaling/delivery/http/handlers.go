package http

import (
	"encoding/json"
	"log"
	"net/http"
	"video-call/internal/models"
	"video-call/internal/signaling"
	"video-call/pkg/logger"

	signalingWs "video-call/internal/signaling/delivery/ws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for chat features
type Handler struct {
	useCase               signaling.UseCase
	wsNotificationHandler *signalingWs.WsNotificationHandler
	logger                logger.Logger
}

// NewHandler creates a new chat HTTP handler
func NewHandler(useCase signaling.UseCase, wsNotificationHandler *signalingWs.WsNotificationHandler, logger logger.Logger) *Handler {
	return &Handler{
		useCase:               useCase,
		wsNotificationHandler: wsNotificationHandler,
		logger:                logger,
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

	notificationPayload := map[string]interface{}{
		"event": "incoming_call",
		"data": map[string]string{
			"callId": call.ID.String(),
			"caller": callerID.String(),
			"callee": calleeID.String(),
		},
	}

	payloadBytes, err := json.Marshal(notificationPayload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal notification payload"})
		return
	}
	err = h.wsNotificationHandler.SendMessageToUser(calleeID, payloadBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		log.Printf("Failed to send notification: %v", err)
		return
	}
	err = h.useCase.UpdateCallStatus(c.Request.Context(), call.ID, models.CallStatusInitiated, models.CallStatusRinging, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update call status"})
		return
	}
	c.JSON(http.StatusOK, callResponse{Call: call, Role: role})
}
