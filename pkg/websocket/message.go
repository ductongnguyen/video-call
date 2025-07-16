package websocket

type BroadcastMessage struct {
	Topic   string
	Payload []byte
	Sender  *Client // The client who sent the message
}
