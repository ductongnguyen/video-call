package usecase

import (
	"context"
	"time"
	"video-call/config"
	"video-call/internal/models"
	"video-call/internal/signaling"
	"video-call/pkg/logger"

	"github.com/google/uuid"
)

// usecase implements the chat.UseCase interface.
type usecase struct {
	cfg    *config.Config
	repo   signaling.Repository
	logger logger.Logger
}

// NewUseCase is the constructor for the chat use case.
func NewUseCase(cfg *config.Config, repo signaling.Repository, logger logger.Logger) signaling.UseCase {
	return &usecase{
		cfg:    cfg,
		repo:   repo,
		logger: logger,
	}
}

func (u *usecase) CreateOrJoinCall(ctx context.Context, userA, userB uuid.UUID) (*models.Call, string, error) {
	if userA == userB {
		return nil, "", signaling.ErrPermissionDenied
	}
	// Chuẩn hóa caller/callee
	callerID, calleeID := userA, userB
	role := "caller"
	if userA.String() > userB.String() {
		callerID, calleeID = userB, userA
		if userA != callerID {
			role = "callee"
		}
	} else if userA != callerID {
		role = "callee"
	}
	// Kiểm tra call active giữa 2 user
	call, err := u.repo.GetActiveByUserPair(ctx, callerID, calleeID)
	if err == nil {
		// Đã có call active, join vào
		return call, role, nil
	}
	if err != signaling.ErrCallNotFound {
		return nil, "", err
	}
	// Tạo call mới
	call = &models.Call{
		CallerID:    callerID,
		CalleeID:    calleeID,
		InitiatedID: userA,
		Status:      models.CallStatusInitiated,
	}
	if err := u.repo.Create(ctx, call); err != nil {
		return nil, "", err
	}
	return call, role, nil
}

func (u *usecase) UpdateCallStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error {
	return u.repo.UpdateStatus(ctx, callID, from, to, answeredAt, endedAt)
}

func (u *usecase) GetCallByID(ctx context.Context, id uuid.UUID) (*models.Call, error) {
	return u.repo.GetByID(ctx, id)
}
