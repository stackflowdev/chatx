package websocket

import (
	"edu-tga/internal/message"
	"time"

	"golang.org/x/net/websocket"
)

// WebSocket connection sozlamalari
const (
	writeWait      = 10 * time.Second    // Browser'ga xabar yozish uchun maksimal kutish vaqti
	pongWait       = 60 * time.Second    // Browser'dan pong javobini kutish vaqti
	pingPeriod     = (pongWait * 9) / 10 // Qancha vaqtda ping yuborish (54 soniya)
	maxMessageSize = 512                 // Maksimal xabar hajmi (bayt'da)
)

// Client - bitta WebSocket connection'ni boshqaradi.
// Har bir ulangan foydalanuvchi uchun alohida Client yaratiladi.
// Client ikkita goroutine'da ishlaydi:
//   - ReadPump: Browser'dan xabar o'qiydi va Hub'ga yuboradi
//   - WritePump: Hub'dan xabar olib, browser'ga yozadi
type Client struct {
	Hub      *Hub                  // Qaysi Hub'ga tegishli (barcha clientlarni boshqaruvchi)
	Conn     *websocket.Conn       // WebSocket connection
	Send     chan *message.Message // Hub'dan kelgan xabarlarni kutish uchun channel (buffered)
	Username string                // Foydalanuvchi ismi
	RoomID   string                // Qaysi xonada (room)
}

// ReadPump - Browser'dan xabarlarni o'qish uchun goroutine.
// Bu method abadiy tsiklda ishlab, WebSocket connection'dan
// JSON formatdagi xabarlarni o'qiydi va Hub'ga tarqatish uchun yuboradi.
//
// Connection uzilganda yoki xato yuz berganda:
//  1. Client'ni Hub'dan unregister qiladi
//  2. WebSocket connection'ni yopadi
func (c *Client) ReadPump() {
	// defer - function tugaganda (xato yoki normal) bajariladi
	defer func() {
		c.Hub.Unregister <- c // Hub'ga "men ketdim" deb xabar
		c.Conn.Close()        // Connection'ni yopish
	}()

	// Abadiy tsikl - xabarlarni o'qish
	for {
		var msg message.Message
		// Browser'dan JSON xabar o'qish (blocking - xabar kelguncha kutadi)
		err := websocket.JSON.Receive(c.Conn, &msg)
		if err != nil {
			// Xato (connection uzilgan, timeout, format xato) - tsikldan chiqish
			break
		}

		// Xabarga metadata qo'shish (client browser'da yubormasligi mumkin)
		msg.Username = c.Username
		msg.RoomID = c.RoomID
		msg.Timestamp = time.Now().Unix()

		// Xabarni Hub'ga yuborish - Hub buni barcha clientlarga tarqatadi
		c.Hub.Broadcast <- &msg
	}
}

// WritePump - Browser'ga xabar yozish uchun goroutine.
// Bu method ikki vazifani bajaradi:
//  1. Hub'dan c.Send channel orqali kelgan xabarlarni browser'ga yozadi
//  2. Har 54 soniyada ping yuboradi (connection'ni alive tutish uchun)
//
// Connection uzilganda yoki channel yopilganda:
//   - Ticker'ni to'xtatadi
//   - WebSocket connection'ni yopadi
func (c *Client) WritePump() {
	// Ping yuborish uchun timer (har 54 soniyada)
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()  // Timer'ni to'xtatish
		c.Conn.Close() // Connection'ni yopish
	}()

	// Abadiy tsikl - xabar yoki ping yuborish
	for {
		select {
		// Case 1: Hub'dan xabar keldi
		case message, ok := <-c.Send:
			if !ok {
				// Channel yopilgan - client unregister qilingan
				return
			}

			// Xabarni JSON formatda browser'ga yuborish
			err := websocket.JSON.Send(c.Conn, message)
			if err != nil {
				// Yozish xatosi - connection uzilgan
				return
			}

		// Case 2: Ping vaqti keldi (har 54 soniyada)
		case <-ticker.C:
			// Ping yuborish - connection alive ekanini tekshirish
			// Oddiy chat uchun bu optional, lekin production'da foydali
			// (Browser pong javob bermasa, connection o'lib qolgan deb hisoblanadi)
		}
	}
}
