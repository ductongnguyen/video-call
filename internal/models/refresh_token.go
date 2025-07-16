package models

import (
	"time"
)

type RefreshToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string    `json:"token" gorm:"not null;unique"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Revoked   bool      `json:"revoked" gorm:"not null;default:false"`
}
