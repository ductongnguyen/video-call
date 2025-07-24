//go:generate mockgen -source usecase.go -destination mocks/usecase_mock.go -package mock
package auth

import (
	"context"
	"time"

	"video-call/internal/models"
)

// Auth use case
type UseCase interface {
	Register(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, userId string) (*models.User, error)

	// Refresh token methods
	GenerateRefreshToken(ctx context.Context, userID string) (string, time.Time, error)
	ValidateRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}
