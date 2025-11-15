package validator

import (
	"errors"
	"strings"
)

// Validation constants - maksimal va minimal uzunliklar.
// Bu qiymatlar butun app'da bir xil bo'lishi uchun constant qilingan.
//
// Nima uchun limit kerak?
//   - Username: 20 belgi etarli, uzun bo'lsa UI'da ko'rinmaydi
//   - RoomID: 30 belgi - aniq va o'qilishi oson nom uchun
//   - MinUsername: kamida 1 belgi - bo'sh username qabul qilinmaydi
//
// Agar kerak bo'lsa, bu qiymatlarni config'ga o'tkazish mumkin.
const (
	MaxUsernameLength = 20 // Username uchun maksimal belgilar soni
	MaxRoomIDLength   = 30 // Room nomi uchun maksimal belgilar soni
	MinUsernameLength = 1  // Username uchun minimal belgilar soni
)

// ValidateUsername - foydalanuvchi nomini tekshiradi.
// Bu funksiya quyidagi validation'larni bajaradi:
//   1. Bo'sh bo'lmasligi (TrimSpace'dan keyin)
//   2. Minimal uzunlik (1 belgi)
//   3. Maksimal uzunlik (20 belgi)
//
// Parametrlar:
//   username - tekshiriladigan foydalanuvchi nomi
//
// Qaytaradi:
//   error - agar validation o'tmasa, xato matni
//   nil - agar barchasi OK bo'lsa
//
// Misol:
//   err := ValidateUsername("Ali")      // nil (OK)
//   err := ValidateUsername("")         // "username bo'sh bo'lishi mumkin emas"
//   err := ValidateUsername("VeryLongNameMoreThan20Chars") // "username juda uzun"
func ValidateUsername(username string) error {
	// TrimSpace - bosh va oxiridagi bo'shliqlarni olib tashlash
	// Misol: "  Ali  " -> "Ali"
	username = strings.TrimSpace(username)

	// Bo'sh tekshiruvi - TrimSpace'dan keyin hali ham bo'sh bo'lishi mumkin
	if username == "" {
		return errors.New("username bo'sh bo'lishi mumkin emas")
	}

	// Minimal uzunlik tekshiruvi (aslida bu yerda ortiqcha, chunki bo'sh emas)
	if len(username) < MinUsernameLength {
		return errors.New("username juda qisqa")
	}

	// Maksimal uzunlik tekshiruvi - spam va UI muammolaridan qochish
	if len(username) > MaxUsernameLength {
		return errors.New("username juda uzun (max 20 belgi)")
	}

	// Barcha tekshiruvlar o'tdi
	return nil
}

// ValidateRoomID - xona (room) nomini tekshiradi.
// Room ID optional - agar berilmasa, "general" default room ishlatiladi.
//
// Validation'lar:
//   1. Bo'sh bo'lishi mumkin (default room ishlatiladi)
//   2. Agar berilgan bo'lsa, maksimal uzunlik (30 belgi)
//
// Parametrlar:
//   roomID - tekshiriladigan room nomi
//
// Qaytaradi:
//   error - agar validation o'tmasa, xato matni
//   nil - agar barchasi OK bo'lsa yoki bo'sh bo'lsa
//
// Misol:
//   err := ValidateRoomID("")          // nil (default room ishlatiladi)
//   err := ValidateRoomID("general")   // nil (OK)
//   err := ValidateRoomID("VeryLongRoomNameMoreThan30Characters") // xato
func ValidateRoomID(roomID string) error {
	// Bo'sh bo'lsa OK - handler'da "general" default qilinadi
	if roomID == "" {
		return nil
	}

	// Bo'shliqlarni olib tashlash
	roomID = strings.TrimSpace(roomID)

	// Maksimal uzunlik tekshiruvi
	if len(roomID) > MaxRoomIDLength {
		return errors.New("room nomi juda uzun (max 30 belgi)")
	}

	return nil
}

// ValidateMessageContent - xabar matnini tekshiradi.
// Bu validation client.ReadPump ichida chat xabarlar uchun ishlatiladi.
//
// Validation'lar:
//   1. Bo'sh bo'lmasligi (TrimSpace'dan keyin)
//   2. Maksimal uzunlik (1000 belgi)
//
// Parametrlar:
//   content - tekshiriladigan xabar matni
//
// Qaytaradi:
//   error - agar validation o'tmasa, xato matni
//   nil - agar barchasi OK bo'lsa
//
// Nima uchun 1000 belgi?
//   - Chat xabar odatda qisqa bo'ladi (bir nechta jumla)
//   - Spam va DoS attack'lardan himoya
//   - WebSocket maxMessageSize (512 bayt) bilan mos emas, lekin
//     bu yerda Unicode belgilar soni (1 belgi = 1-4 bayt bo'lishi mumkin)
//
// Misol:
//   err := ValidateMessageContent("Salom")  // nil (OK)
//   err := ValidateMessageContent("")       // "xabar matni bo'sh bo'lishi mumkin emas"
//   err := ValidateMessageContent(strings.Repeat("a", 1001)) // "xabar juda uzun"
func ValidateMessageContent(content string) error {
	// Bo'shliqlarni olib tashlash
	content = strings.TrimSpace(content)

	// Bo'sh xabar qabul qilinmaydi
	if content == "" {
		return errors.New("xabar matni bo'sh bo'lishi mumkin emas")
	}

	// Maksimal uzunlik - spam prevention
	if len(content) > 1000 {
		return errors.New("xabar juda uzun (max 1000 belgi)")
	}

	return nil
}
