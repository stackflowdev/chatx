package handler

import (
	"chatx/internal/config"
	"chatx/internal/message"
	"chatx/internal/validator"
	"chatx/internal/websocket"
	"log"
	"net/http"
	"time"

	ws "golang.org/x/net/websocket"
)

// ChatHandler - WebSocket connection'larni boshqarish uchun HTTP handler.
// Hub bilan bog'langan va har bir yangi WebSocket connection uchun
// Client yaratadi va Hub'ga register qiladi.
type ChatHandler struct {
	hub *websocket.Hub // Barcha clientlarni boshqaruvchi Hub
	cfg *config.Config
}

// NewChatHandler - yangi ChatHandler instance yaratadi.
// hub parametri orqali dependency injection qilinadi.
func NewChatHandler(hub *websocket.Hub, cfg *config.Config) *ChatHandler {
	return &ChatHandler{
		hub: hub,
		cfg: cfg,
	}
}

// ServeWS - WebSocket connection'ni boshqarish uchun HTTP handler method.
//
// Bu method quyidagi ishlarni bajaradi:
//  1. HTTP request'dan username va room parametrlarini oladi
//  2. HTTP connection'ni WebSocket'ga upgrade qiladi
//  3. Yangi Client yaratadi va Hub'ga register qiladi
//  4. Join message yuboradi (barcha clientlarga)
//  5. ReadPump va WritePump goroutine'larini ishga tushiradi
//
// URL format: ws://localhost:8080/ws?username=Ali&room=general
func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket request keldi: %s", r.URL.String())

	username := r.URL.Query().Get("username")
	roomID := r.URL.Query().Get("room")

	if err := validator.ValidateUsername(username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Username validation xatosi: %v", err)
		return
	}

	if err := validator.ValidateRoomID(roomID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Room validation xatosi: %v", err)
		return
	}

	if roomID == "" {
		roomID = "general"
	}

	// WebSocket server yaratish va HTTP connection'ni WebSocket'ga upgrade qilish
	// ws.Server - golang.org/x/net/websocket kutubxonasidan
	server := ws.Server{
		// Handler - WebSocket connection ochilgandan keyin chaqiriladi
		Handler: func(conn *ws.Conn) {
			log.Printf("WebSocket ulandi: %s, xona: %s", username, roomID)

			// Yangi Client yaratish - bu foydalanuvchining WebSocket connection'i
			client := &websocket.Client{
				Hub:      h.hub,                            // Qaysi Hub'ga tegishli
				Conn:     conn,                             // WebSocket connection
				Send:     make(chan *message.Message, 256), // Xabar yuborish uchun buffered channel
				Username: username,                         // Foydalanuvchi ismi
				RoomID:   roomID,                           // Xona nomi
			}

			log.Printf("Client yaratildi: %s, xona: %s", username, roomID)

			// Client'ni Hub'ga register qilish
			// Hub endi bu clientni ro'yxatida saqlaydi va xabar tarqatadi
			client.Hub.Register <- client

			// Message history yuborish - yangi user ulanganda o'sha room'ning eski xabarlarini ko'radi
			// Oxirgi 50 ta xabarni store'dan olib, faqat shu client'ga yuborish
			history := client.Hub.Store.GetRecentMessages(roomID, h.cfg.HistoryCount)
			for _, msg := range history {
				client.Send <- msg // Faqat bu client'ga (boshqalarga emas)
			}

			// Join message yaratish va barcha clientlarga yuborish
			// "Ali has joined the room" kabi xabar
			joinMsg := &message.Message{
				Type:      message.MessageTypeJoin,
				Content:   username + " has joined the room.",
				Username:  username,
				RoomID:    roomID,
				Timestamp: time.Now().Unix(),
			}

			// Join message'ni Hub'ga yuborish - Hub buni barcha clientlarga tarqatadi
			client.Hub.Broadcast <- joinMsg

			// Ikkita goroutine ishga tushirish:
			// 1. WritePump: Hub'dan xabar olib, browser'ga yozadi (alohida goroutine)
			// 2. ReadPump: Browser'dan xabar o'qiydi va Hub'ga yuboradi (blocking)
			go client.WritePump()
			client.ReadPump() // Bu method qaytmaguncha handler tugamaydi
		},
	}

	// HTTP request'ni WebSocket'ga upgrade qilish va handler'ni ishga tushirish
	server.ServeHTTP(w, r)
}
