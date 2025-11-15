package main

import (
	"context"
	"chatx/internal/config"
	"chatx/internal/handler"
	"chatx/internal/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main - dasturning asosiy kirish nuqtasi.
// Bu yerda:
//  1. Hub yaratiladi va alohida goroutine'da ishga tushiriladi
//  2. HTTP handler'lar sozlanadi
//  3. Server port 8080'da ishga tushadi
func main() {
	// 1) Config o'qish
	cfg := config.NewConfig()
	log.Printf("Starting server with config: port=%s store_max=%d history=%d",
		cfg.ServerPort, cfg.StoreMaxSize, cfg.HistoryCount)

	// 2) Hub yaratish (store size config orqali uzatiladi)
	hub := websocket.NewHub(cfg)

	// 3) Hub'ni kontekst bilan ishga tushirish (graceful stop uchun)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// 4) Handler yaratish va route sozlash
	// Handler'ga hub va config uzatiladi (dependency injection pattern)
	chatHandler := handler.NewChatHandler(hub, cfg)
	http.HandleFunc("/ws", chatHandler.ServeWS)

	// 5) HTTP server yaratish va alohida goroutine'da ishga tushirish
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
