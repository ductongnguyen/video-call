package signaling

import (
	"context"
	"time"
	"video-call/internal/models"

	"github.com/google/uuid"
)

type UseCase interface {
	CreateOrJoinCall(ctx context.Context, userA, userB uuid.UUID) (*models.Call, string, error)
	UpdateCallStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error
	GetCallByID(ctx context.Context, id uuid.UUID) (*models.Call, error)
}
