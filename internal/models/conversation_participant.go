package models

import "time"

// ConversationParticipant represents the conversation_participants table
type ConversationParticipant struct {
	ConversationID string    `json:"conversation_id" gorm:"type:char(36);primaryKey"`
	UserID         string    `json:"user_id" gorm:"type:char(36);primaryKey"`
	JoinedAt       time.Time `json:"joined_at"`
}
