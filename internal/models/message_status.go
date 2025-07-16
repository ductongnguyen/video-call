package models

import "time"

// MessageStatus represents the message_status table
type MessageStatus struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
