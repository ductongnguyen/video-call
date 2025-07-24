package signaling

import (
    "context"
    "github.com/google/uuid"
    "video-call/internal/models"
    "time"
)

type CallUsecase interface {
    // Tạo hoặc join call giữa 2 user, trả về call và vai trò (caller/callee)
    CreateOrJoinCall(ctx context.Context, userA, userB uuid.UUID) (*models.Call, string, error)
    // Cập nhật trạng thái cuộc gọi
    UpdateCallStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error
    // Lấy thông tin call theo ID
    GetCallByID(ctx context.Context, id uuid.UUID) (*models.Call, error)
} 