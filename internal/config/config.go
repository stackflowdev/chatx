package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort     string
	StoreMaxSize   int
	HistoryCount   int
	MaxMessageSize int

	// ENV: CHATX_WRITE_WAIT (masalan "10s")
	WriteWait time.Duration

	// PongWait - Browser'dan pong javobini kutish uchun maksimal vaqt.
	// Server ping yuboradi, browser pong bilan javob berishi kerak.
	// Agar bu vaqtda javob bo'lmasa, connection o'lik deb hisoblanadi.
	// ENV: CHATX_PONG_WAIT (masalan "60s")
	PongWait time.Duration

	// PingPeriod - qancha vaqt oralig'ida ping yuborilishi.
	// Odatda PongWait * 9/10 bo'ladi (54 soniya agar PongWait=60s).
	// ENV: CHATX_PING_PERIOD (masalan "54s")
	PingPeriod time.Duration

	LogLevel string
}

func NewConfig() *Config {
	// Default qiymatlar - env berilmasa ishlatiladi
	const (
		defPort           = ":8080"
		defStoreMaxSize   = 100
		defHistoryCount   = 50
		defMaxMessageSize = 512
		defWriteWait      = 10 * time.Second
		defPongWait       = 60 * time.Second
		defPingPeriod     = 54 * time.Second
		defLogLevel       = "info"
	)

	cfg := &Config{
		ServerPort:     getEnv("CHATX_PORT", defPort),
		StoreMaxSize:   mustAtoi(os.Getenv("CHATX_STORE_MAX"), defStoreMaxSize),
		HistoryCount:   mustAtoi(os.Getenv("CHATX_HISTORY_COUNT"), defHistoryCount),
		MaxMessageSize: mustAtoi(os.Getenv("CHATX_MAX_MESSAGE_SIZE"), defMaxMessageSize),
		WriteWait:      mustDuration(os.Getenv("CHATX_WRITE_WAIT"), defWriteWait),
		PongWait:       mustDuration(os.Getenv("CHATX_PONG_WAIT"), defPongWait),
		PingPeriod:     mustDuration(os.Getenv("CHATX_PING_PERIOD"), defPingPeriod),
		LogLevel:       getEnv("CHATX_LOG_LEVEL", defLogLevel),
	}

	return cfg
}

func getEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustAtoi(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

func mustDuration(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}
