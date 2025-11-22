package websocket

import (
	"chatx/internal/config"
	"chatx/internal/message"
	"chatx/internal/store"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// Hub - barcha WebSocket client'larni boshqaruvchi markaz (dispatcher).
// Vazifasi:
// 1. Yangi client'larni ro'yxatga olish (register)
// 2. Uzilib qolgan client'larni o'chirish (unregister)
// 3. Xabarlarni barcha client'larga tarqatish (broadcast)
//
// Hub alohida goroutine'da ishlaydi va channel'lar orqali client'lar bilan
// thread-safe aloqa qiladi. Bu pattern'ga "hub-and-spoke" deyiladi.
//
// Nima uchun Hub kerak?
//   - Barcha client'larni markazdan boshqarish (centralized control)
//   - Thread-safety: faqat 1 ta goroutine Clients map'ini o'zgartiradi
//   - Message broadcasting: 1 ta xabar kelsa, barchasiga yuboradi
//   - Room filtering: har bir room alohida xabar oladi
//
// Nima uchun channel'lar?
//   - Go'da channel'lar thread-safe (mutex kerak emas)
//   - Goroutine'lar o'rtasida xavfsiz aloqa
//   - select statement bilan bir nechta event'ni kutish mumkin
type Hub struct {
	// Clients - hozirda ulangan barcha client'lar ro'yxati.
	// Key: Client pointer, Value: true (mavjudligini bildiradi)
	// Map ishlatilishi sababi: tez qidirish va o'chirish (O(1) complexity)
	Clients map[*Client]bool

	// Broadcast - client'lardan kelgan xabarlarni barcha client'larga
	// tarqatish uchun channel. Client xabar yozsa, bu channel'ga tushadi
	// va Hub uni barcha ulangan client'larga yuboradi.
	Broadcast chan *message.Message

	// Register - yangi client ulanganda, uni ro'yxatga olish uchun channel.
	// Handler yangi client yaratganda, bu channel'ga yuboradi va
	// Hub uni Clients map'iga qo'shadi.
	Register chan *Client

	// Unregister - client uzilganda, uni ro'yxatdan o'chirish uchun channel.
	// Client connection uzilganda yoki xato bo'lganda, ReadPump defer'da
	// bu channel'ga yuboradi va Hub uni Clients map'idan o'chiradi.
	Unregister chan *Client

	// Store - xabarlarni saqlash uchun in-memory storage.
	// Yangi user ulanganda oxirgi xabarlarni ko'rsatish imkonini beradi.
	Store *store.Store

	// Config - ilova konfiguratsiyasi (timeouts, sizes, history count va boshqalar).
	// Hub orqali barcha client'lar config'ga murojaat qilishi mumkin.
	Config      *config.Config
	OnlineUsers map[string]map[string]time.Time // roomID -> username -> last seen time
	mu          sync.Mutex
}

// Nima uchun make() ishlatiladi?
//   - make(map) - bo'sh map yaratadi (nil emas!)
//   - make(chan) - channel yaratadi (buffered yoki unbuffered)
//   - Agar make() qilmasak, map va channel nil bo'ladi va panic beradi
func NewHub(cfg *config.Config) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan *message.Message),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Store:       store.NewStore(cfg.StoreMaxSize),
		Config:      cfg,
		OnlineUsers: make(map[string]map[string]time.Time),
	}
}

// Run - Hub'ning asosiy ishlash sikli. Bu method alohida goroutine'da
// ishga tushirilishi kerak: go hub.Run()
//
// Bu method abadiy tsiklda 3 ta channel'ni tinglaydi:
// - Register: yangi client ulanganda
// - Unregister: client uzilganda
// - Broadcast: xabar tarqatish kerak bo'lganda
//
// Select statement orqali qaysi channel'ga ma'lumot kelganini aniqlaydi
// va tegishli amalni bajaradi.
func (h *Hub) Run(ctx context.Context) {
	cleanupTicker := time.NewTicker(30 * time.Second)
	defer cleanupTicker.Stop()

	for {
		select {

		case <-ctx.Done():
			log.Printf("Hub: context done, shutting down")
			// Qo'shimcha cleanup (connectionlarni yoping) kerak bo'lsa shu yerga qo'shing
			return

		case <-cleanupTicker.C:
			h.cleanupStaleUsers()

		case client := <-h.Register:
			h.Clients[client] = true
			// Online users'ga qo'shish
			h.mu.Lock()
			if h.OnlineUsers[client.RoomID] == nil {
				h.OnlineUsers[client.RoomID] = make(map[string]time.Time)
			}
			h.OnlineUsers[client.RoomID][client.Username] = time.Now()
			h.mu.Unlock()

			// Presence broadcast
			h.broadcastPresence(client.RoomID)

			log.Printf("Hub: register client %s (room=%s)", client.Username, client.RoomID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("Hub: unregister client %s (room=%s)", client.Username, client.RoomID)

				// Online users'dan o'chirish
				h.mu.Lock()
				if h.OnlineUsers[client.RoomID] != nil {
					delete(h.OnlineUsers[client.RoomID], client.Username)
				}
				h.mu.Unlock()

				h.broadcastPresence(client.RoomID)

				// Leave message yaratish va barcha qolgan clientlarga yuborish
				leaveMsg := &message.Message{
					Type:      message.MessageTypeLeave,
					Content:   client.Username + " has left the room.",
					Username:  client.Username,
					RoomID:    client.RoomID,
					Timestamp: time.Now().Unix(),
				}

				// Faqat o'sha room'dagi clientlarga yuborish
				for c := range h.Clients {
					if c.RoomID != client.RoomID {
						continue
					}
					select {
					case c.Send <- leaveMsg:
					default:
						close(c.Send)
						delete(h.Clients, c)
					}
				}
			}

		case msg := <-h.Broadcast:
			// Xabarni store'ga saqlash (faqat 1 marta, room bo'yicha)
			// Typing message'lar tarixda saqlanmaydi
			if h.Store != nil && msg != nil && msg.Type != message.MessageTypeTyping {
				h.Store.AddMessage(msg)
			}

			// Faqat o'sha room'dagi clientlarga yuborish
			for client := range h.Clients {
				// Room filterlash - faqat bir xil room'dagi clientlar xabar oladi
				if client.RoomID == msg.RoomID {
					select {
					case client.Send <- msg:
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
		}
	}
}

// broadcastPresence - room'dage online user'lar ro'xatini broadcast qiladi
func (h *Hub) broadcastPresence(roomID string) {
	h.mu.Lock()
	users := make([]string, 0, len(h.OnlineUsers[roomID]))
	for username := range h.OnlineUsers[roomID] {
		users = append(users, username)
	}
	h.mu.Unlock()

	usersJSON, _ := json.Marshal(users)

	presenceMsg := &message.Message{
		Type:      message.MessageTypePresence,
		Content:   string(usersJSON),
		RoomID:    roomID,
		Timestamp: time.Now().Unix(),
	}

	// Faqat o'sha room'dagi clientlarge yuborish
	for client := range h.Clients {
		if client.RoomID == roomID {
			select {
			case client.Send <- presenceMsg:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}

// cleanupStaleUsers - 2 minutdan ortiq ko'rinmagan user'larni online ro'yxatidan o'chiradi
func (h *Hub) cleanupStaleUsers() {
	now := time.Now()
	staleThreshold := 2 * time.Minute
	h.mu.Lock()
	defer h.mu.Unlock()

	for roomID, users := range h.OnlineUsers {
		for username, lastActivity := range users {
			if now.Sub(lastActivity) > staleThreshold {
				delete(users, username)
				log.Printf("HUB: cleanup stale user %s from room %s", username, roomID)
				// Presence update yuborish (Lock ichida emas)
				go h.broadcastPresence(roomID)
			}
		}
	}
}
