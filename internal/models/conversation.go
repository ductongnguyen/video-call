package models

import (
	"time"

	"github.com/google/uuid"
)

// Conversation represents the conversations table
type Conversation struct {
	ID        string    `json:"id" gorm:"type:char(36);primary_key"`
	UUID      uuid.UUID `json:"uuid" gorm:"type:char(36);unique"`
	IsGroup   bool      `json:"is_group"`
	Name      *string   `json:"name"`
	CreatedBy string    `json:"created_by" gorm:"type:char(36)"`
	CreatedAt time.Time `json:"created_at"`
}
