package main

import (
	"chatx/internal/config"
	"chatx/internal/handler"
	"chatx/internal/websocket"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.NewConfig()
	log.Printf("Starting server with config: port=%s history=%d", cfg.ServerPort, cfg.HistoryCount)

	hub := websocket.NewHub(cfg)

	// Hub'ni kontekst bilan ishga tushirish (graceful stop uchun)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Handler'ga hub va config uzatiladi (dependency injection pattern)
	chatHandler := handler.NewChatHandler(hub, cfg)
	http.HandleFunc("/ws", chatHandler.ServeWS)

	// HTTP server yaratish va alohida goroutine'da ishga tushirish
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: nil,
	}

	go func() {
		log.Printf("Server listening on %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// 6) Signal kutish (SIGINT/SIGTERM) -> graceful shutdown
	// SIGINT - Ctrl+C bosilganda
	// SIGTERM - kill command yoki orchestrator (Docker, K8s) yuborilganda
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	<-sigC // Signal kelguncha bloklangan holda kutadi
	log.Println("Shutdown signal received, starting graceful shutdown...")

	// Hub'ni to'xtatish - context'ni bekor qilish orqali
	// Hub.Run() ichidagi <-ctx.Done() case'i ishga tushadi
	cancel()

	// HTTP serverga graceful shutdown uchun 10 soniya vaqt berish
	// Bu vaqt ichida:
	//   1. Yangi request'lar qabul qilinmaydi
	//   2. Mavjud request'lar tugashini kutadi
	//   3. Agar 10 soniyada tugamasa, majburan to'xtatadi
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}
