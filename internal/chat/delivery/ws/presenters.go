package delivery

import (
	"github.com/ductongnguyen/vivy-chat/internal/models"
	"gorm.io/datatypes"
)

// CreateMessageRequest defines the expected structure for incoming WebSocket messages.
type CreateMessageRequest struct {
	Content     string             `json:"content" validate:"required"`
	MessageType models.MessageType `json:"message_type" validate:"required"`
	Metadata    datatypes.JSON     `json:"metadata,omitempty"`
}
