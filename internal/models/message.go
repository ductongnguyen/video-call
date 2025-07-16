package models

import (
	"time"

	"gorm.io/datatypes"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeVideo MessageType = "video"
	MessageTypeFile  MessageType = "file"
)

// Message represents the messages table
type Message struct {
	ID             string         `gorm:"primaryKey;type:char(36)" json:"id"`
	ConversationID string         `json:"conversation_id" gorm:"type:char(36)"`
	SenderID       string         `json:"sender_id" gorm:"type:char(36)"`
	Content        string         `json:"content"`
	MessageType    MessageType    `json:"message_type"`
	Metadata       datatypes.JSON `json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}
