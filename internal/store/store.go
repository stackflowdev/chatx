package store

import (
	"chatx/internal/message"
	"sync"
)

// Store - xabarlarni in-memory (xotirada) room'lar bo'yicha saqlash uchun struktura.
// Bu database o'rniga oddiy xotiradan foydalanadi - yangi xabar kelganda
// saqlaydi va yangi user ulanganda o'sha room'ning eski xabarlarini ko'rsatadi.
//
// Thread-safe: Ko'p goroutine bir vaqtda ishlatishi mumkin, RWMutex orqali
// himoyalangan. Bu chat app'da juda muhim, chunki ko'p client bir vaqtda
// xabar yuborishi va o'qishi mumkin.
//
// Room-based: Har bir room uchun alohida xabarlar ro'yxati saqlanadi.
// Masalan: "general" room'ining xabarlari "golang" room'idan ajratilgan.
type Store struct {
	mu       sync.RWMutex                  // Read-Write qulf - thread-safe qilish uchun
	messages map[string][]*message.Message // Key: roomID, Value: xabarlar slice'i
	maxSize  int                           // Har bir room uchun maksimal xabarlar soni
}

func NewStore(maxSize int) *Store {
	return &Store{
		messages: make(map[string][]*message.Message), // Bo'sh map - room'lar kerak bo'lganda yaratiladi
		maxSize:  maxSize,
	}
}

// AddMessage - yangi xabarni store'ga qo'shadi.
// Thread-safe: Lock bilan himoyalangan, faqat 1 ta goroutine yozishi mumkin.
//
// Agar xabarlar soni maxSize'dan oshsa, eng eski xabar o'chiriladi.
// Misol: maxSize=100 bo'lsa, 101-chi xabar kelganda 1-chi xabar o'chadi.
func (s *Store) AddMessage(msg *message.Message) {
	s.mu.Lock()         // Qulfni yopish - faqat men yozaman
	defer s.mu.Unlock() // Function tugaganda qulfni ochish
	roomId := msg.RoomID

	if s.messages[roomId] == nil {
		s.messages[roomId] = make([]*message.Message, 0, s.maxSize)
	}

	s.messages[roomId] = append(s.messages[roomId], msg)

	// Agar limit oshib ketsa, eng eski xabarni o'chirish
	if len(s.messages[roomId]) > s.maxSize {
		s.messages[roomId] = s.messages[roomId][1:] // 1-chisini o'chirish, qolganlarni siljitish
	}
}

// GetRecentMessages - oxirgi N ta xabarni qaytaradi.
// count - nechta xabar kerak (masalan oxirgi 50 ta).
//
// Thread-safe: RLock bilan himoyalangan, ko'p goroutine bir vaqtda
// o'qishi mumkin (lekin yozish bo'lmaydi).
//
// Agar count store'dagi xabarlardan ko'p bo'lsa, barcha xabarlar qaytariladi.
// Agar store bo'sh bo'lsa yoki count <= 0 bo'lsa, bo'sh slice qaytariladi.
func (s *Store) GetRecentMessages(roomID string, count int) []*message.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roomMessages := s.messages[roomID]

	if len(roomMessages) == 0 || count <= 0 {
		return []*message.Message{}
	}

	if count > len(roomMessages) {
		count = len(roomMessages)
	}

	return roomMessages[len(roomMessages)-count:]
}
