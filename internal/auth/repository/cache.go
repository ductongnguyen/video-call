package repository

import (
	"context"
	"encoding/json"

	"github.com/ductongnguyen/vivy-chat/internal/auth"
	"github.com/ductongnguyen/vivy-chat/internal/models"
	"github.com/ductongnguyen/vivy-chat/pkg/cache/redis"
)

// News redis repository
type redisRepo struct {
	rdb redis.Client
}

// News redis repository constructor
func NewRedisRepo(rdb redis.Client) auth.RedisRepository {
	return &redisRepo{rdb: rdb}
}

// GetUserByIDCtx implements auth.RedisRepository.
func (r *redisRepo) GetUserByIDCtx(ctx context.Context, key string) (*models.User, error) {

	data, err := r.rdb.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var user models.User                      // Declare a variable to hold the unmarshalled data
	err = json.Unmarshal([]byte(data), &user) // Unmarshal the byte slice representation of the string
	if err != nil {
		return nil, err
	}
	return &user, nil // Return the unmarshalled user object
}

// SetUserByIDCtx implements auth.RedisRepository.
func (r *redisRepo) SetUserByIDCtx(ctx context.Context, key string, user *models.User) error {
	data, err := json.Marshal(user) // Marshal the user object to a JSON string
	if err != nil {
		return err
	}
	err = r.rdb.Set(ctx, key, string(data), 0) // Store the JSON string in Redis with no expiration time
	if err != nil {
		return err
	}
	return nil // Return nil if no error occurred
}
