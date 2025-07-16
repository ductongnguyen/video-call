package models

import (
    "time"
    "github.com/google/uuid"
)

type CallStatus string

const (
    CallStatusInitiated CallStatus = "initiated"
    CallStatusRinging   CallStatus = "ringing"
    CallStatusActive    CallStatus = "active"
    CallStatusEnded     CallStatus = "ended"
    CallStatusRejected  CallStatus = "rejected"
    CallStatusMissed    CallStatus = "missed"
)

type Call struct {
    ID           uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
    CallerID     uuid.UUID   `gorm:"type:uuid;not null" json:"caller_id"`
    CalleeID     uuid.UUID   `gorm:"type:uuid;not null" json:"callee_id"`
    InitiatedID  uuid.UUID   `gorm:"type:uuid;not null" json:"initiated_id"`
    Status       CallStatus  `gorm:"type:call_status;not null" json:"status"`
    InitiatedAt  time.Time   `gorm:"type:timestamptz;not null;default:now()" json:"initiated_at"`
    AnsweredAt   *time.Time  `gorm:"type:timestamptz" json:"answered_at,omitempty"`
    EndedAt      *time.Time  `gorm:"type:timestamptz" json:"ended_at,omitempty"`
} 