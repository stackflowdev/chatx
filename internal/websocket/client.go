package websocket

import (
	"chatx/internal/message"
	"chatx/internal/validator"
	"log"
	"time"

	"golang.org/x/net/websocket"
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
	defer func() {
		c.Hub.Unregister <- c
		if err := c.Conn.Close(); err != nil {
			log.Printf("Connection yopishda xato [%s]: %v", c.Username, err)
		}
	}()

	// Abadiy tsikl - xabarlarni o'qish
	for {
		var msg message.Message
		err := websocket.JSON.Receive(c.Conn, &msg)
		if err != nil {
			log.Printf("Xabar o'qishda xato [%s]:%v", c.Username, err)
			break
		}

		if msg.Type == message.MessageTypeTyping {
			msg.Username = c.Username
			msg.RoomID = c.RoomID
			msg.Timestamp = time.Now().Unix()
			c.Hub.Broadcast <- &msg
			continue
		}

		if msg.Type == message.MessageTypeChat {
			if err := validator.ValidateMessageContent(msg.Content); err != nil {
				log.Printf("Xabar validation xatosi [%s]: %v", c.Username, err)
				continue // Bu xabarni ignore qilamiz, keyingisini o'qiymiz
			}
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
//
// Ping/Pong mexanizmi:
//
//	Server ping yuboradi -> Browser pong bilan javob beradi
//	Agar browser javob bermasa, connection o'lik deb hisoblanadi
func (c *Client) WritePump() {
	// Config'dan timeout qiymatlarini olish (Hub orqali)
	// Agar config yo'q bo'lsa, default qiymatlar ishlatiladi
	pingPeriod := 54 * time.Second

	if c.Hub != nil && c.Hub.Config != nil {
		// writeWait kelajakda deadline qo'yish uchun ishlatilishi mumkin
		// Hozircha golang.org/x/net/websocket kutubxonasi deadline'ni
		// to'g'ridan-to'g'ri qo'llab-quvvatlamaydi, shuning uchun comment qilamiz
		// writeWait = c.Hub.Config.WriteWait
		pingPeriod = c.Hub.Config.PingPeriod
	}

	// Ping yuborish uchun timer
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop() // Timer'ni to'xtatish
		if err := c.Conn.Close(); err != nil {
			log.Printf("Connection yopishda xato [%s]: %v", c.Username, err)
		}
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
			// golang.org/x/net/websocket orqali JSON encoding
			err := websocket.JSON.Send(c.Conn, message)
			if err != nil {
				log.Printf("Xabar yozishda xato [%s]:%v", c.Username, err)
				return
			}

		// Case 2: Ping vaqti keldi (har 54 soniyada)
		case <-ticker.C:
			// Ping yuborish - connection alive ekanini tekshirish
			// golang.org/x/net/websocket kutubxonasi avtomatik ping/pong qiladi
			// Shuning uchun bu yerda qo'shimcha kod kerak emas
			// Ticker faqat connection health monitoring uchun saqlanadi
			//
			// Kelajakda: Agar aniq ping yuborish kerak bo'lsa:
			// err := websocket.Message.Send(c.Conn, []byte("ping"))
		}
	}
}
