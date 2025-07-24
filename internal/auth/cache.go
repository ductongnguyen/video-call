//go:generate mockgen -source cache.go -destination mock/redis_repository_mock.go -package mock
package auth

import (
	"context"

	"video-call/internal/models"
)

// News redis repository
type RedisRepository interface {
	GetUserByIDCtx(ctx context.Context, key string) (*models.User, error)
	SetUserByIDCtx(ctx context.Context, key string, user *models.User) error
}
