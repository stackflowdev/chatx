package store

import (
	"chatx/internal/message"
	"sync"
)

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

func (s *Store) AddMessage(msg *message.Message) {
	s.mu.Lock()         // Qulfni yopish - faqat men yozaman
	defer s.mu.Unlock() // Function tugaganda qulfni ochish
	roomId := msg.RoomID

	if s.messages[roomId] == nil {
		s.messages[roomId] = make([]*message.Message, 0, s.maxSize)
	}

	s.messages[roomId] = append(s.messages[roomId], msg)

	if len(s.messages[roomId]) > s.maxSize {
		s.messages[roomId] = s.messages[roomId][1:]
	}
}

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
