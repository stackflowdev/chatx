package main

import (
	"edu-tga/internal/handler"
	"edu-tga/internal/websocket"
	"log"
	"net/http"
)

// main - dasturning asosiy kirish nuqtasi.
// Bu yerda:
//  1. Hub yaratiladi va alohida goroutine'da ishga tushiriladi
//  2. HTTP handler'lar sozlanadi
//  3. Server port 8080'da ishga tushadi
func main() {
	// 1. Hub yaratish va alohida goroutine'da ishga tushirish
	// Hub - barcha WebSocket clientlarni boshqaruvchi markaz
	hub := websocket.NewHub()
	go hub.Run() // Alohida goroutine'da abadiy ishlab turadi

	// 2. HTTP handler yaratish
	// ChatHandler - HTTP requestlarni qabul qilib, WebSocket'ga upgrade qiladi
	chatHandler := handler.NewChatHandler(hub)

	// 3. Route'lar sozlash
	// Root endpoint - server ishlayotganini tekshirish uchun
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Chat Server ishlayapti!\n"))
	})
	// WebSocket endpoint - chat uchun asosiy endpoint
	http.HandleFunc("/ws", chatHandler.ServeWS)

	// 4. Server ishga tushirish
	addr := ":8080"
	log.Printf("Server ishga tushdi: http://localhost%s", addr)
	log.Printf("Websocket endpoint: ws://localhost%s/ws?username=Name&room=general", addr)

	// HTTP server'ni ishga tushirish (blocking - server to'xtatilguncha kutadi)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server xatosi:", err)
	}
}
