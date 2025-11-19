package websocket

import (
	"chatx/internal/config"
	"chatx/internal/message"
	"chatx/internal/store"
	"context"
	"log"
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
	Config *config.Config
}

// NewHub - yangi Hub instance yaratadi va barcha zarur channel'lar hamda
// map'larni initialize qiladi. Bu constructor pattern - Go'da struct'larni
// to'g'ri boshlang'ich holatda yaratish uchun ishlatiladigan standart usul.
//
// Parametrlar:
//
//	cfg - ilova konfiguratsiyasi (store size, timeouts va boshqalar)
//
// Qaytaradi:
//
//	*Hub - to'liq initialize qilingan Hub instance
//
// Nima uchun make() ishlatiladi?
//   - make(map) - bo'sh map yaratadi (nil emas!)
//   - make(chan) - channel yaratadi (buffered yoki unbuffered)
//   - Agar make() qilmasak, map va channel nil bo'ladi va panic beradi
func NewHub(cfg *config.Config) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *message.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Store:      store.NewStore(cfg.StoreMaxSize),
		Config:     cfg, // Config'ni saqlash - client'lar ishlatishi uchun
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
	for {
		select {

		case <-ctx.Done():
			log.Printf("Hub: context done, shutting down")
			// Qo'shimcha cleanup (connectionlarni yoping) kerak bo'lsa shu yerga qo'shing
			return

		case client := <-h.Register:
			h.Clients[client] = true
			log.Printf("Hub: register client %s (room=%s)", client.Username, client.RoomID)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("Hub: unregister client %s (room=%s)", client.Username, client.RoomID)

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
