package websocket

// Message types - xabar turlari
// Bu konstantlar client va server o'rtasida yuborilayotgan
// xabar turini aniqlash uchun ishlatiladi
const (
	MessageTypeChat   = "chat"   // Oddiy chat xabari
	MessageTypeJoin   = "join"   // Foydalanuvchi xonaga kirdi
	MessageTypeLeave  = "leave"  // Foydalanuvchi xonadan chiqdi
	MessageTypeTyping = "typing" // Foydalanuvchi yozyapti (keyingi feature)
)

// Message - client va server o'rtasida almashiladigan asosiy xabar strukturasi.
// JSON formatda serialize/deserialize qilinadi.
type Message struct {
	Type      string `json:"type"`              // Xabar turi (chat, join, leave, typing)
	Content   string `json:"content,omitempty"` // Xabar matni (omitempty - bo'sh bo'lsa JSON'ga qo'shilmaydi)
	Username  string `json:"username"`          // Xabar yuborgan foydalanuvchi ismi
	RoomID    string `json:"room_id"`           // Qaysi xonaga tegishli
	Timestamp int64  `json:"timestamp"`         // Xabar yuborilgan vaqt (Unix timestamp)
}
