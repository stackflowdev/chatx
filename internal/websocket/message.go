package websocket

// Message types
const (
	MessageTypeChat   = "chat"
	MessageTypeJoin   = "join"
	MessageTypeLeave  = "leave"
	MessageTypeTyping = "typing"
)

type Message struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Username  string `json:"username"`
	RoomID    string `json:"room_id"`
	Timestamp int64  `json:"timestamp"`
}
