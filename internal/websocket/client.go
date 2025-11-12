package websocket

type Client struct {
	// Buffered channel of outbound messages
	send chan *Message
}
