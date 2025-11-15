package config

import (
	"os"
	"strconv"
	"time"
)

// Config - ilova konfiguratsiyasi uchun asosiy struktura.
// Bu struct environment variable'lar orqali sozlanadi, agar ular
// berilmasa default qiymatlar ishlatiladi.
//
// Maqsad: Barcha "magic numbers" va sozlamalarni bir joyda to'plash.
// Bu development va production muhitlari o'rtasida oson o'tish,
// testing'da sozlamalarni o'zgartirish va maintainability uchun muhim.
//
// Misol:
//
//	Development: Default qiymatlar (port :8080, kichik store size)
//	Production: ENV orqali sozlash (EDU_TGA_PORT=:80, katta store size)
type Config struct {
	// ServerPort - HTTP server qaysi port'da tinglashi kerak.
	// Masalan ":8080" yoki ":443" (HTTPS uchun).
	// ENV: CHATX_PORT
	ServerPort string

	// StoreMaxSize - har bir room uchun in-memory store'da saqlanadigan
	// maksimal xabarlar soni. Agar bu limit oshsa, eng eski xabar o'chiriladi.
	// Katta qiymat - ko'proq xotira ishlatiladi, kichik - kam tarix.
	// ENV: CHATX_STORE_MAX
	StoreMaxSize int

	// HistoryCount - yangi user ulanganda nechta eski xabar yuborilishi.
	// Masalan 50 - oxirgi 50 ta xabarni ko'radi.
	// ENV: CHATX_HISTORY_COUNT
	HistoryCount int

	// MaxMessageSize - WebSocket orqali qabul qilinadigan maksimal xabar hajmi (bayt).
	// Bu spam va DoS attack'lardan himoya qilish uchun.
	// ENV: CHATX_MAX_MESSAGE_SIZE
	MaxMessageSize int

	// WriteWait - Browser'ga xabar yozish uchun maksimal kutish vaqti.
	// Agar bu vaqtda yozilmasa, connection uzilgan deb hisoblanadi.
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

	// LogLevel - log'lash darajasi ("info", "debug", "error").
	// Development'da "debug", production'da "info" bo'lishi mumkin.
	// ENV: CHATX_LOG_LEVEL
	LogLevel string
}

// NewConfig - yangi Config instance yaratadi va environment variable'lardan
// yoki default qiymatlardan to'ldiradi.
//
// Environment variable'lar:
//
//	CHATX_PORT - server port (default: ":8080")
//	CHATX_STORE_MAX - store size (default: 100)
//	CHATX_HISTORY_COUNT - history count (default: 50)
//	CHATX_MAX_MESSAGE_SIZE - max message size (default: 512 bytes)
//	CHATX_WRITE_WAIT - write timeout (default: "10s")
//	CHATX_PONG_WAIT - pong timeout (default: "60s")
//	CHATX_PING_PERIOD - ping period (default: "54s")
//	CHATX_LOG_LEVEL - log level (default: "info")
//
// Ishlash usuli:
//  1. Env variable borligini tekshiradi
//  2. Agar bor bo'lsa, parse qiladi (string, int, duration)
//  3. Agar yo'q bo'lsa yoki parse xatosi bo'lsa, default ishlatadi
func NewConfig() *Config {
	// Default qiymatlar - env berilmasa ishlatiladi
	const (
		defPort           = ":8080"          // HTTP server default port
		defStoreMaxSize   = 100              // Har bir room uchun 100 ta xabar
		defHistoryCount   = 50               // Yangi user uchun 50 ta xabar
		defMaxMessageSize = 512              // 512 bayt (qisqa xabarlar)
		defWriteWait      = 10 * time.Second // 10 soniya write timeout
		defPongWait       = 60 * time.Second // 60 soniya pong kutish
		defPingPeriod     = 54 * time.Second // 54 soniya ping oralig'i
		defLogLevel       = "info"           // Info level logging
	)

	// Config yaratish va env'dan to'ldirish
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

// getEnv - environment variable'ni o'qiydi va agar bo'sh bo'lsa default qaytaradi.
// Bu string qiymatlar uchun ishlatiladi (masalan port, log level).
//
// Parametrlar:
//
//	key - env variable nomi (masalan "CHATX_PORT")
//	def - default qiymat (agar env bo'lmasa)
//
// Misol:
//
//	port := getEnv("CHATX_PORT", ":8080")
//	Agar CHATX_PORT set bo'lmasa, ":8080" qaytariladi
func getEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// mustAtoi - string'ni int'ga o'zgartiradi, xato bo'lsa default qaytaradi.
// "must" prefixi Go'da "panic qilmaydi, default qaytaradi" degan ma'noni bildiradi.
//
// Parametrlar:
//
//	s - parse qilinadigan string (masalan "100")
//	def - default qiymat (agar parse xatosi bo'lsa)
//
// Misol:
//
//	size := mustAtoi("100", 50)  // 100 qaytaradi
//	size := mustAtoi("abc", 50)  // 50 qaytaradi (parse xatosi)
//	size := mustAtoi("", 50)     // 50 qaytaradi (bo'sh string)
func mustAtoi(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// mustDuration - string'ni time.Duration'ga o'zgartiradi, xato bo'lsa default qaytaradi.
// Duration format: "10s", "5m", "1h" (s=soniya, m=minut, h=soat).
//
// Parametrlar:
//
//	s - parse qilinadigan string (masalan "10s")
//	def - default qiymat (agar parse xatosi bo'lsa)
//
// Misol:
//
//	timeout := mustDuration("10s", 5*time.Second)  // 10 soniya
//	timeout := mustDuration("abc", 5*time.Second)  // 5 soniya (parse xatosi)
//	timeout := mustDuration("", 5*time.Second)     // 5 soniya (bo'sh string)
func mustDuration(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}
