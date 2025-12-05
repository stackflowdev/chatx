package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort     string
	StoreMaxSize   int
	HistoryCount   int
	MaxMessageSize int
	WriteWait      time.Duration
	// PongWait - Browser'dan pong javobini kutish uchun maksimal vaqt.
	// Server ping yuboradi, browser pong bilan javob berishi kerak.
	// Agar bu vaqtda javob bo'lmasa, connection o'lik deb hisoblanadi.
	// ENV: CHATX_PONG_WAIT (masalan "60s")
	PongWait time.Duration

	// PingPeriod - qancha vaqt oralig'ida ping yuborilishi.
	// Odatda PongWait * 9/10 bo'ladi (54 soniya agar PongWait=60s).
	// ENV: CHATX_PING_PERIOD (masalan "54s")
	PingPeriod time.Duration
	LogLevel   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Connection pool sozlamalari - Go database/sql package uchun
	// DBMaxOpenConns - maksimal ochiq connection'lar soni
	// Agar barcha connection'lar band bo'lsa, yangi so'rov kutadi
	DBMaxOpenConns int

	// DBMaxIdleConns - maksimal idle (ishlatilmayotgan) connection'lar
	// Idle connection'larni qayta ishlatish - yangi connection ochishdan tezroq
	DBMaxIdleConns int

	// DBConnMaxLifetime - connection maksimal umri
	// Eski connection'larni yopish va yangisini ochish (memory leak'dan qochish)
	DBConnMaxLifetime time.Duration
}

func NewConfig() *Config {
	const (
		defPort           = ":8080"
		defStoreMaxSize   = 100
		defHistoryCount   = 50
		defMaxMessageSize = 512
		defWriteWait      = 10 * time.Second
		defPongWait       = 60 * time.Second
		defPingPeriod     = 54 * time.Second
		defLogLevel       = "info"

		defDBHost            = "localhost"
		defDBPort            = "5432"
		defDBUser            = "chatuser"
		defDBPassword        = "chatpass"
		defDBName            = "chatx"
		defDBSSLMode         = "disable"
		defDBMaxOpenConns    = 25
		defDBMaxIdleConns    = 5
		defDBConnMaxLifetime = 5 * time.Minute
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

		DBHost:            getEnv("CHATX_DB_HOST", defDBHost),
		DBPort:            getEnv("CHATX_DB_PORT", defDBPort),
		DBUser:            getEnv("CHATX_DB_USER", defDBUser),
		DBPassword:        getEnv("CHATX_DB_PASSWORD", defDBPassword),
		DBName:            getEnv("CHATX_DB_NAME", defDBName),
		DBSSLMode:         getEnv("CHATX_DB_SSLMODE", defDBSSLMode),
		DBMaxOpenConns:    mustAtoi(os.Getenv("CHATX_DB_MAX_OPEN_CONNS"), defDBMaxOpenConns),
		DBMaxIdleConns:    mustAtoi(os.Getenv("CHATX_DB_MAX_IDLE_CONNS"), defDBMaxIdleConns),
		DBConnMaxLifetime: mustDuration(os.Getenv("CHATX_DB_CONN_MAX_LIFETIME"), defDBConnMaxLifetime),
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

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}
