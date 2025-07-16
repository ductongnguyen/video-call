package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	redis "github.com/redis/go-redis/v9"
)

type HubIface interface {
	Register(client *Client)
	Unregister(client *Client)
}

type RedisHub struct {
	clients          map[*Client]bool
	clientsMutex     sync.RWMutex
	redisClient      *redis.Client
	ctx              context.Context
	subscribedTopics map[string]bool
	subscribedMutex  sync.Mutex
}

func NewRedisHub(redisClient *redis.Client) *RedisHub {
	return &RedisHub{
		clients:          make(map[*Client]bool),
		redisClient:      redisClient,
		ctx:             context.Background(),
		subscribedTopics: make(map[string]bool),
	}
}

func (h *RedisHub) Register(client *Client) {
	h.clientsMutex.Lock()
	h.clients[client] = true
	h.clientsMutex.Unlock()
}

func (h *RedisHub) Unregister(client *Client) {
	h.clientsMutex.Lock()
	delete(h.clients, client)
	fmt.Printf("[Unregister] client.ID=%v, client=%p\n", client.ID, client)
	h.clientsMutex.Unlock()
}

// Subscribe vào topic Redis và forward message cho client local
func (h *RedisHub) SubscribeTopic(topic string) {
	h.subscribedMutex.Lock()
	if h.subscribedTopics[topic] {
		h.subscribedMutex.Unlock()
		return // Already subscribed
	}
	h.subscribedTopics[topic] = true
	h.subscribedMutex.Unlock()

	go func() {
		pubsub := h.redisClient.Subscribe(h.ctx, topic)
		ch := pubsub.Channel()
		for msg := range ch {
			// Forward cho tất cả client local đang theo dõi topic này
			h.clientsMutex.RLock()
			for client := range h.clients {
				if client.Topics[topic] {
					// Kiểm tra sender_id trong message với client.ID
					var payload map[string]interface{}
					if err := json.Unmarshal([]byte(msg.Payload), &payload); err == nil {
						senderID := payload["sender_id"]
						senderIDStr := fmt.Sprintf("%v", senderID)
						clientIDStr := fmt.Sprintf("%v", client.ID)
						if senderIDStr == clientIDStr {
							continue // Bỏ qua client gửi
						}
					}
					client.send <- []byte(msg.Payload)
				}
			}
			h.clientsMutex.RUnlock()
		}
	}()
}

// Publish message lên Redis
func (h *RedisHub) Publish(topic string, message []byte) error {
	return h.redisClient.Publish(h.ctx, topic, message).Err()
}
