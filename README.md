# ChatX - Real-time WebSocket Chat Application

Go tilida yozilgan, minimal dependency'lar bilan real-time chat ilovasi. Educational project - Go concurrency pattern'larini o'rganish uchun.

## Features
- ✅ WebSocket Real-time Chat - bir vaqtda xabar almashish
- ✅ Hub-and-Spoke Pattern - markazlashtirilgan connection management
- ✅ Goroutines & Channels - concurrent programming
- ✅ Thread-safe Storage - Mutex bilan xavfsiz xotira
- ✅ Message History - yangi user eski xabarlarni ko'radi
- ✅ Multiple Rooms - turli xonalar, alohida history
- ✅ Join/Leave Messages - kim keldi/ketdi xabarlari
- ✅ Configuration Management - env variable'lar bilan sozlash
- ✅ Graceful Shutdown - context bilan to'g'ri to'xtash
- ✅ Input Validation - username, room, message validation

## Tech Stack
- **Language:** Go 1.25+
- **WebSocket:** golang.org/x/net/websocket
- **Architecture:** Hub-and-Spoke pattern
- **Storage:** In-memory (map + mutex)
- **Config:** Environment variables

## Project Structure
```
chatx/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── config/config.go         # Configuration management
│   ├── handler/handler.go       # HTTP -> WebSocket upgrade
│   ├── message/message.go       # Message struct
│   ├── store/store.go           # In-memory storage
│   ├── validator/validator.go   # Input validation
│   └── websocket/
│       ├── hub.go               # Connection manager
│       └── client.go            # Per-connection handler
├── .env                         # Configuration file
└── test-client.html             # Browser test client
```

## Quick Start

1. **Clone va dependencies:**
```bash
git clone <repo>
cd chatx
go mod download
```

2. **Serverni ishga tushirish:**
```bash
go run cmd/server/main.go
```

3. **Browser'da test qilish:**
```bash
open test-client.html
```

## Configuration

`.env` fayl yarating yoki environment variable'lar bilan sozlang:

```bash
# ChatX Application Configuration
CHATX_PORT=:8080
CHATX_STORE_MAX=100
CHATX_HISTORY_COUNT=50
CHATX_MAX_MESSAGE_SIZE=512
CHATX_WRITE_WAIT=10s
CHATX_PONG_WAIT=60s
CHATX_PING_PERIOD=54s
CHATX_LOG_LEVEL=info
```

**Env bilan ishga tushirish:**
```bash
CHATX_PORT=:8081 CHATX_STORE_MAX=200 go run cmd/server/main.go
```

## Go'da o'rganganlar
- **Concurrency:** goroutines, channels, select statement
- **Patterns:** Hub-and-Spoke, Constructor, Dependency Injection
- **Thread-safety:** sync.RWMutex, Lock/Unlock
- **Network:** WebSocket, HTTP handlers
- **Architecture:** Package separation, import cycles hal qilish
- **Configuration:** Environment variables, default values
- **Context:** Graceful shutdown, cancellation
- **Validation:** Input validation, error handling
- **Data structures:** Maps, slices, channels

## WebSocket API

**Connect:**
```
ws://localhost:8080/ws?username=Ali&room=general
```

**Message format (JSON):**
```json
{
  "type": "chat",
  "content": "Salom hammaga!",
  "username": "Ali",
  "room_id": "general",
  "timestamp": 1234567890
}
```

**Message types:**
- `chat` - Oddiy chat xabari
- `join` - User xonaga kirdi
- `leave` - User xonadan chiqdi

## Testing

**Build:**
```bash
go build ./...
```

**Test with multiple rooms:**
1. Browser'da bir nechta tab oching
2. Har xil username va room bilan ulaning
3. Xabar yuboring va room filter'ni tekshiring

**Graceful shutdown:**
```bash
Ctrl+C  # yoki kill -TERM <pid>
```

## Keyingi qadamlar (To-Do)
- [ ] Unit Tests - validator, store, hub testlari
- [ ] Rate Limiting - spam prevention
- [ ] Metrics - monitoring va statistics
- [ ] User List - kimlar online ko'rsatish
- [ ] Typing Indicator - kim yozyapti
- [ ] Private Messages - shaxsiy xabarlar
- [ ] Authentication - JWT token
- [ ] Database - PostgreSQL persistence
- [ ] Docker - containerization
- [ ] Deploy - production'ga chiqarish

## License
Educational project - free to use and learn from.