package repository

import (
	"video-call/internal/chat"
	"video-call/pkg/cache/redis"
)

// redisRepo implements the chat.RedisRepository interface.
type redisRepo struct {
	rdb redis.Client
}

// NewRedisRepo is the constructor for redisRepo.
func NewRedisRepo(rdb redis.Client) chat.RedisRepository {
	return &redisRepo{rdb: rdb}
}
