package repository

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
    "video-call/internal/models"
    "video-call/internal/signaling"
)

type postgresRepo struct {
    db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) signaling.CallRepository {
    return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(ctx context.Context, call *models.Call) error {
    return r.db.WithContext(ctx).Create(call).Error
}

func (r *postgresRepo) UpdateStatus(ctx context.Context, callID uuid.UUID, from, to models.CallStatus, answeredAt, endedAt *time.Time) error {
    updates := map[string]interface{}{"status": to}
    if answeredAt != nil {
        updates["answered_at"] = *answeredAt
    }
    if endedAt != nil {
        updates["ended_at"] = *endedAt
    }
    tx := r.db.WithContext(ctx).Model(&models.Call{}).
        Where("id = ? AND status = ?", callID, from).
        Updates(updates)
    if tx.Error != nil {
        return tx.Error
    }
    if tx.RowsAffected == 0 {
        return signaling.ErrInvalidTransition
    }
    return nil
}

func (r *postgresRepo) GetActiveByUserPair(ctx context.Context, userA, userB uuid.UUID) (*models.Call, error) {
    var call models.Call
    callerID, calleeID := userA, userB
    if userA.String() > userB.String() {
        callerID, calleeID = userB, userA
    }
    err := r.db.WithContext(ctx).
        Where("caller_id = ? AND callee_id = ? AND status IN ?", callerID, calleeID, []models.CallStatus{
            models.CallStatusInitiated, models.CallStatusRinging, models.CallStatusActive,
        }).
        First(&call).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, signaling.ErrCallNotFound
    }
    return &call, err
}

func (r *postgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Call, error) {
    var call models.Call
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&call).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, signaling.ErrCallNotFound
    }
    return &call, err
} 