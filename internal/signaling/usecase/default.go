package usecase

import (
    "context"
    "github.com/google/uuid"
    "video-call/internal/models"
    "video-call/internal/signaling"
    "time"
)

type defaultUsecase struct {
    repo signaling.CallRepository
}

func NewCallUsecase(repo signaling.CallRepository) signaling.CallUsecase {
    return &defaultUsecase{repo: repo}
}

func (u *defaultUsecase) CreateOrJoinCall(ctx context.Context, userA, userB uuid.UUID) (*models.Call, string, error) {
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

func (u *defaultUsecase) UpdateCallStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error {
    return u.repo.UpdateStatus(ctx, callID, from, to, answeredAt, endedAt)
}

func (u *defaultUsecase) GetCallByID(ctx context.Context, id uuid.UUID) (*models.Call, error) {
    return u.repo.GetByID(ctx, id)
} 