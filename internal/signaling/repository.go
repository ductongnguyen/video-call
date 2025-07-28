package signaling

import (
	"context"
	"time"
	"video-call/internal/models"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, call *models.Call) error
	UpdateStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error
	GetActiveByUserPair(ctx context.Context, userA, userB uuid.UUID) (*models.Call, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Call, error)
}
