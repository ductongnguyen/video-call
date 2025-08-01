package http

import (
	"net/http"
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
	roomID := uuid.New().String()
	c.JSON(http.StatusOK, callResponse{RoomID: roomID})
}
