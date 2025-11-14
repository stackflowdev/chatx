package store

import (
	"edu-tga/internal/message"
	"sync"
)

// Store - xabarlarni in-memory (xotirada) saqlash uchun struktura.
// Bu database o'rniga oddiy xotiradan foydalanadi - yangi xabar kelganda
// saqlaydi va yangi user ulanganda eski xabarlarni ko'rsatish imkonini beradi.
//
// Thread-safe: Ko'p goroutine bir vaqtda ishlatishi mumkin, RWMutex orqali
// himoyalangan. Bu chat app'da juda muhim, chunki ko'p client bir vaqtda
// xabar yuborishi va o'qishi mumkin.
type Store struct {
	mu       sync.RWMutex       // Read-Write qulf - thread-safe qilish uchun
	messages []*message.Message // Saqlangan barcha xabarlar (slice)
	maxSize  int                // Maksimal saqlash hajmi (eski xabarlar o'chiriladi)
}

// NewStore - yangi Store instance yaratadi.
// maxSize - maksimal nechta xabar saqlanishi (masalan 100).
// Agar limit to'lsa, eng eski xabar o'chiriladi (FIFO - First In First Out).
func NewStore(maxSize int) *Store {
	return &Store{
		messages: make([]*message.Message, 0, maxSize), // Bo'sh slice, capacity = maxSize
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

	s.messages = append(s.messages, msg) // Xabarni slice'ga qo'shish

	// Agar limit oshib ketsa, eng eski xabarni o'chirish
	if len(s.messages) > s.maxSize {
		s.messages = s.messages[1:] // 1-chisini o'chirish, qolganlarni siljitish
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
func (s *Store) GetRecentMessages(count int) []*message.Message {
	s.mu.RLock()         // O'qish qulfi - ko'pchilik o'qishi mumkin
	defer s.mu.RUnlock() // Function tugaganda qulfni ochish

	// Agar messages bo'sh yoki count noto'g'ri bo'lsa
	if len(s.messages) == 0 || count <= 0 {
		return []*message.Message{} // Bo'sh slice qaytarish
	}

	// Agar count juda katta bo'lsa, barcha xabarlarni qaytarish
	if count > len(s.messages) {
		count = len(s.messages)
	}

	// Oxirgi 'count' ta xabarni slice qilish va qaytarish
	// Masalan: agar 100 ta xabar bor va count=50, oxirgi 50 tasini olish
	return s.messages[len(s.messages)-count:]
}
