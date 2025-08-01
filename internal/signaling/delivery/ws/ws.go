package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WsNotificationHandler struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewWsNotificationHandler() *WsNotificationHandler {
	return &WsNotificationHandler{
		rooms: make(map[string]*Room),
	}
}

func (h *WsNotificationHandler) GetOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomID]; ok {
		return room
	}

	room := NewRoom(roomID, h)
	h.rooms[roomID] = room
	go room.Run()
	log.Printf("Room created: %s", roomID)
	return room
}

func (h *WsNotificationHandler) removeRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[roomID]; ok {
		delete(h.rooms, roomID)
		log.Printf("Room removed as it became empty: %s", roomID)
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Cho phép tất cả các nguồn gốc
	},
}

// ServeWs xử lý các yêu cầu websocket từ người dùng.
func (h *WsNotificationHandler) ServeWs(c *gin.Context) {
	roomID := c.Query("roomId")
	userID := c.Query("userId")
	fmt.Println(roomID, userID)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	room := h.GetOrCreateRoom(roomID)
	participant := NewParticipant(userID, room, conn)

	room.register <- participant // Đăng ký người tham gia mới vào phòng.

	go participant.writePump()
	go participant.readPump()
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 10
)

// Participant là một người dùng được kết nối vào một Room.
type Participant struct {
	userID string // Có thể dùng string hoặc uuid.UUID tùy vào yêu cầu
	room   *Room
	conn   *websocket.Conn
	send   chan []byte // Kênh chứa các tin nhắn gửi đi
}

func NewParticipant(userID string, room *Room, conn *websocket.Conn) *Participant {
	for p := range room.participants {
		if p.userID == userID {
			delete(room.participants, p)
		}
	}
	return &Participant{
		userID: userID,
		room:   room,
		conn:   conn,
		send:   make(chan []byte, 256),
	}
}

// readPump đọc tin nhắn từ kết nối websocket và gửi đến phòng.
func (p *Participant) readPump() {
	defer func() {
		p.room.unregister <- p
		p.conn.Close()
	}()
	p.conn.SetReadLimit(maxMessageSize)
	p.conn.SetReadDeadline(time.Now().Add(pongWait))
	p.conn.SetPongHandler(func(string) error {
		p.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error from %s: %v", p.userID, err)
			}
			break
		}
		// Đóng gói tin nhắn cùng với thông tin người gửi và đưa vào kênh broadcast của phòng
		p.room.broadcast <- &BroadcastMessage{sender: p, payload: message}
	}
}

// writePump ghi tin nhắn từ phòng ra kết nối websocket.
func (p *Participant) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.conn.Close()
	}()
	for {
		select {
		case message, ok := <-p.send:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				p.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := p.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			p.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// BroadcastMessage là một tin nhắn cần được phát sóng, chứa thông tin người gửi.
type BroadcastMessage struct {
	sender  *Participant
	payload []byte
}

type Room struct {
	id           string
	hub          *WsNotificationHandler
	participants map[*Participant]bool
	broadcast    chan *BroadcastMessage
	register     chan *Participant
	unregister   chan *Participant
}

func NewRoom(id string, hub *WsNotificationHandler) *Room {
	return &Room{
		id:           id,
		hub:          hub,
		participants: make(map[*Participant]bool),
		broadcast:    make(chan *BroadcastMessage),
		register:     make(chan *Participant),
		unregister:   make(chan *Participant),
	}
}

func (r *Room) Run() {
	for {
		select {
		case participant := <-r.register:
			r.participants[participant] = true
			log.Printf("Participant %s joined room %s. Total: %d", participant.userID, r.id, len(r.participants))
			r.notifyParticipantJoined(participant)

		case participant := <-r.unregister:
			delete(r.participants, participant)
			close(participant.send)
			log.Printf("Participant %s left room %s. Total: %d", participant.userID, r.id, len(r.participants))
			r.notifyParticipantLeft(participant)
			if len(r.participants) == 0 {
				r.hub.removeRoom(r.id) // Tự hủy phòng nếu trống
				return
			}

		case message := <-r.broadcast:
			r.handleBroadcast(message)
		}
	}
}

type ServerMessage struct {
	Event    string      `json:"event"`
	SenderID string      `json:"senderId,omitempty"` // ID của người gửi gốc
	Data     interface{} `json:"data"`
}

// handleBroadcast xử lý một tin nhắn đến và chuyển tiếp nó đến những người khác.
func (r *Room) handleBroadcast(msg *BroadcastMessage) {
	var clientMsg map[string]interface{}
	if err := json.Unmarshal(msg.payload, &clientMsg); err != nil {
		log.Printf("Could not parse message from %s: %v", msg.sender.userID, err)
		return
	}

	// Server chỉ quan tâm đến việc đóng gói lại và gửi đi
	// Client sẽ chịu trách nhiệm về `event` và `data`
	finalPayload, _ := json.Marshal(ServerMessage{
		Event:    clientMsg["event"].(string),
		SenderID: msg.sender.userID, // Luôn đính kèm ID người gửi
		Data:     clientMsg["data"],
	})

	for participant := range r.participants {
		// Gửi cho tất cả mọi người TRỪ người gửi
		if participant.userID != msg.sender.userID {
			select {
			case participant.send <- finalPayload:
			default:
				close(participant.send)
				delete(r.participants, participant)
			}
		}
	}
}

// Gửi thông báo có người mới tham gia đến tất cả những người khác trong phòng.
func (r *Room) notifyParticipantJoined(joinedParticipant *Participant) {
	// Lấy danh sách ID của tất cả người tham gia hiện tại
	allParticipantIDs := make([]string, 0)
	for p := range r.participants {
		if p.userID != joinedParticipant.userID {
			allParticipantIDs = append(allParticipantIDs, p.userID)
		}
	}

	notification, _ := json.Marshal(ServerMessage{
		Event: "participant-joined",
		Data: map[string]string{
			"joinedId": joinedParticipant.userID,
		},
	})

	// Gửi thông báo người mới vào cho những người cũ
	for p := range r.participants {
		if p.userID != joinedParticipant.userID {
			p.send <- notification
		}
	}

	// Gửi thông báo người mới vào cho người mới
	welcomeNotification, _ := json.Marshal(ServerMessage{
		Event: "room-joined",
		Data: map[string]interface{}{
			"roomId":       r.id,
			"participants": allParticipantIDs,
		},
	})
	joinedParticipant.send <- welcomeNotification
}

// Gửi thông báo có người rời đi đến những người còn lại.
func (r *Room) notifyParticipantLeft(leftParticipant *Participant) {
	notification, _ := json.Marshal(ServerMessage{
		Event: "participant-left",
		Data:  map[string]string{"leftId": leftParticipant.userID},
	})

	for participant := range r.participants {
		select {
		case participant.send <- notification:
		default:
			close(participant.send)
			delete(r.participants, participant)
		}
	}
}
