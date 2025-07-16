package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	WRITEWAIT      = 10 * time.Second
	PONGWAIT       = 60 * time.Second
	PINGPERIOD     = (PONGWAIT * 9) / 10
	MAXMESSAGESIZE = 1024
)

// Client là một middleman giữa kết nối websocket và hub.
type Client struct {
	// ID là một định danh duy nhất cho client (ví dụ: UUID hoặc UserID dạng string).
	ID string

	// Hub mà client này thuộc về.
	hub HubIface

	// Kết nối websocket.
	conn *websocket.Conn

	// Channel đệm cho các message gửi đi.
	send chan []byte

	// Các topic mà client này đã đăng ký.
	Topics map[string]bool

	// OnMessage is a callback function that is called when a message is received from the client.
	OnMessage func(message []byte)
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(MAXMESSAGESIZE)

	// Set the initial read deadline.
	_ = c.conn.SetReadDeadline(time.Now().Add(PONGWAIT))

	// The pong handler resets the read deadline, which is how we know the connection is still alive.
	c.conn.SetPongHandler(func(string) error {
		// A pong was received, so reset the read deadline.
		// This is the direct use of pongWait that satisfies the linter.
		return c.conn.SetReadDeadline(time.Now().Add(PONGWAIT))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if c.OnMessage != nil {
			c.OnMessage(message)
		}
	}
}

// writePump ghi message từ hub ra kết nối websocket.
func (c *Client) WritePump() {
	ticker := time.NewTicker(PINGPERIOD)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(WRITEWAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// NewClient tạo một client mới. Handler sẽ chịu trách nhiệm tạo và đăng ký nó.
func NewClient(hub HubIface, conn *websocket.Conn, id string, onMessage func(message []byte)) *Client {
	return &Client{
		ID:        id,
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		Topics:    make(map[string]bool),
		OnMessage: onMessage,
	}
}
