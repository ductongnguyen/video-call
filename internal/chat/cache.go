//go:generate mockgen -source cache.go -destination mock/redis_repository_mock.go -package mock
package chat

// RedisRepository defines the interface for chat-related cache operations.
type RedisRepository interface {
}
