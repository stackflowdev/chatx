package validator

import (
	"errors"
	"strings"
)

const (
	MaxUsernameLength = 20 // Username uchun maksimal belgilar soni
	MaxRoomIDLength   = 30 // Room nomi uchun maksimal belgilar soni
	MinUsernameLength = 2  // Username uchun minimal belgilar soni
)

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.New("username bo'sh bo'lishi mumkin emas")
	}

	if len(username) < MinUsernameLength {
		return errors.New("username juda qisqa")
	}

	if len(username) > MaxUsernameLength {
		return errors.New("username juda uzun (max 20 belgi)")
	}

	return nil
}

func ValidateRoomID(roomID string) error {
	if roomID == "" {
		return nil
	}

	roomID = strings.TrimSpace(roomID)

	if len(roomID) > MaxRoomIDLength {
		return errors.New("room nomi juda uzun (max 30 belgi)")
	}

	return nil
}

func ValidateMessageContent(content string) error {
	content = strings.TrimSpace(content)

	if content == "" {
		return errors.New("xabar matni bo'sh bo'lishi mumkin emas")
	}

	if len(content) > 1000 {
		return errors.New("xabar juda uzun (max 1000 belgi)")
	}

	return nil
}
