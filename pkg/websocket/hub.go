// pkg/websocket/hub.go
package websocket

// subscriptionRequest dùng để đóng gói yêu cầu đăng ký/hủy đăng ký topic.
type subscriptionRequest struct {
	client *Client
	topic  string
}

// Hub quản lý tập hợp các client đang hoạt động và broadcast message đến các client.
type Hub struct {
	// Map các client đã đăng ký.
	clients map[*Client]bool

	// Map các phòng (topic) và các client trong đó.
	rooms map[string]map[*Client]bool

	// Channel nhận message để broadcast.
	broadcast chan *BroadcastMessage

	// Channel nhận yêu cầu đăng ký client mới.
	register chan *Client

	// Channel nhận yêu cầu hủy đăng ký client.
	unregister chan *Client

	// Channel nhận yêu cầu đăng ký topic.
	subscribe chan *subscriptionRequest

	// Channel nhận yêu cầu hủy đăng ký topic.
	unsubscribe chan *subscriptionRequest
}

func NewHub() *Hub {
	return &Hub{
		broadcast:   make(chan *BroadcastMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		subscribe:   make(chan *subscriptionRequest),
		unsubscribe: make(chan *subscriptionRequest),
		clients:     make(map[*Client]bool),
		rooms:       make(map[string]map[*Client]bool),
	}
}

// Run khởi chạy Hub trong một goroutine riêng.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				// Hủy đăng ký client khỏi tất cả các topic
				for topic := range client.Topics {
					h.handleUnsubscribe(client, topic)
				}
				delete(h.clients, client)
				close(client.send)
			}

		case req := <-h.subscribe:
			h.handleSubscribe(req.client, req.topic)

		case req := <-h.unsubscribe:
			h.handleUnsubscribe(req.client, req.topic)

		case message := <-h.broadcast:
			if room, ok := h.rooms[message.Topic]; ok {
				for client := range room {
					if client != message.Sender {
						select {
						case client.send <- message.Payload:
						default:
							// If the send channel is full, assume the client is lagging and close the connection.
							h.handleUnsubscribe(client, message.Topic)
							delete(h.clients, client)
							close(client.send)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) handleSubscribe(client *Client, topic string) {
	if _, ok := h.rooms[topic]; !ok {
		h.rooms[topic] = make(map[*Client]bool)
	}
	h.rooms[topic][client] = true
	client.Topics[topic] = true
}

func (h *Hub) handleUnsubscribe(client *Client, topic string) {
	if room, ok := h.rooms[topic]; ok {
		delete(room, client)
		delete(client.Topics, topic)
		// Nếu phòng rỗng, xóa phòng để giải phóng bộ nhớ.
		if len(room) == 0 {
			delete(h.rooms, topic)
		}
	}
}

// --- Public API của Hub ---

// Register đăng ký một client với Hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister hủy đăng ký một client khỏi Hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Subscribe đăng ký một client vào một topic.
func (h *Hub) Subscribe(client *Client, topic string) {
	h.subscribe <- &subscriptionRequest{client, topic}
}

// Unsubscribe hủy đăng ký một client khỏi một topic.
func (h *Hub) Unsubscribe(client *Client, topic string) {
	h.unsubscribe <- &subscriptionRequest{client, topic}
}

// Broadcast gửi một message đến tất cả các client trong một topic.
func (h *Hub) Broadcast(message *BroadcastMessage) {
	h.broadcast <- message
}
