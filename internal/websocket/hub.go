package websocket

import (
	"edu-tga/internal/message"
	"edu-tga/internal/store"
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
}

// NewHub - yangi Hub instance yaratadi va barcha zarur channel'lar hamda
// map'larni initialize qiladi. Bu constructor pattern - Go'da struct'larni
// to'g'ri boshlang'ich holatda yaratish uchun ishlatiladigan standart usul.
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *message.Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Store:      store.NewStore(10),
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
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			log.Printf("Client unregister qilinmoqda: %s", client.Username)
			if _, ok := h.Clients[client]; ok {
				// Client'ni ro'yxatdan o'chirish
				delete(h.Clients, client)
				close(client.Send)

				// Leave message yaratish va barcha qolgan clientlarga yuborish
				leaveMsg := &message.Message{
					Type:      message.MessageTypeLeave,
					Content:   client.Username + " has left the room.",
					Username:  client.Username,
					RoomID:    client.RoomID,
					Timestamp: time.Now().Unix(),
				}

				// Leave message'ni barcha qolgan clientlarga yuborish
				for c := range h.Clients {
					select {
					case c.Send <- leaveMsg:
						log.Printf("Leave message yuborildi: %s -> %s", client.Username, c.Username)
					default:
						// Client band bo'lsa, o'tkazib yuboramiz
					}
				}
			}
		case message := <-h.Broadcast:
			// Xabarni store'ga saqlash (faqat 1 marta, room bo'yicha)
			h.Store.AddMessage(message)

			// Faqat o'sha room'dagi clientlarga yuborish
			for client := range h.Clients {
				// Room filterlash - faqat bir xil room'dagi clientlar xabar oladi
				if client.RoomID == message.RoomID {
					select {
					case client.Send <- message:
						// Message muvaffaqiyatli yuborildi
					default:
						// Mijoz javob bermayotganda uni ro'yxatdan o'chirish
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
		}
	}
}
