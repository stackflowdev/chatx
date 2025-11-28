# ChatX - Claude uchun Loyiha Konteksti

## ❗ MUHIM: Claude uchun Ko'rsatmalar

### O'zbek Tilida Tushuntirish Qoidasi
**Bu loyiha Go backend o'rganish uchun yaratilgan educational project.**

Sen (Claude) **DOIM** quyidagicha javob berishing KERAK:
- **Asosiy til:** O'zbek tili (lotin alifbosi)
- **Keywords:** Ingliz tilida qoldiriladi (masalan: `goroutine`, `channel`, `mutex`, `defer`, `select`, `struct`)
- **Code:** Go dasturlash tilida
- **Comments in code:** O'zbek tilida, batafsil tushuntirish bilan

### Tushuntirish Uslubi
Har bir kod yoki tushuntirishda:
1. **Nima qilayotganini** aniq aytish
2. **Nima uchun** shunday qilayotganini tushuntirish
3. **Qanday ishlashini** concurrency va Go pattern'lari bilan bog'lab berish
4. **Misol:** Foydalanuvchiga amaliy misol ko'rsatish

### Misol Format
```go
// ❌ Yomon (qisqa, tushuntirishsiz)
// Client yaratish
client := &Client{...}

// ✅ Yaxshi (batafsil, o'zbek tili + ingliz keywords)
// Client struct'ini yaratish - bu bitta WebSocket connection'ni ifodalaydi
// Har bir ulangan foydalanuvchi uchun alohida Client yaratiladi
// Send channel 256 ta xabarni buffer qiladi (blocking'dan qochish uchun)
client := &Client{
    Hub:      hub,
    Conn:     conn,
    Send:     make(chan *message.Message, 256),
    Username: username,
    RoomID:   roomID,
}
```

---

## Loyiha Haqida Umumiy Ma'lumot

**ChatX** - bu Go tilida yozilgan real-time WebSocket chat ilovasi. Bu loyiha **Go backend o'rganish** maqsadida yaratilgan va Go'ning `concurrency` pattern'larini, `goroutine`lar, `channel`lar va `Hub-and-Spoke` arxitekturasini chuqur o'rgatadi.

### Asosiy O'rganish Maqsadlari
- Go'da `goroutine` va `channel` bilan ishlashni o'rganish
- `Concurrency` va `thread-safety` konseptlarini amaliyotda qo'llash
- `Hub-and-Spoke` pattern'ini tushunish
- `WebSocket` protokoli bilan backend yaratish
- `sync.Mutex` bilan shared resource'larni himoyalash
- `context` bilan graceful shutdown qilish

---

## Texnologiyalar (Tech Stack)

- **Til:** Go 1.25+
- **WebSocket Library:** `golang.org/x/net/websocket`
- **Arxitektura:** Hub-and-Spoke pattern
- **Storage:** In-memory (map + mutex)
- **Konfiguratsiya:** Environment variable'lar (`.env` fayl orqali)

---

## Asosiy Funksiyalar (Features)

- ✅ **Real-time chat** - bir nechta xonalarda (room) bir vaqtda xabar almashish
- ✅ **Direct Messaging (DM)** - `@username` sintaksisi bilan shaxsiy xabar
- ✅ **User Presence Tracking** - kimlar online ekanini ko'rsatish
- ✅ **Typing Indicator** - kim yozyapti ko'rsatish
- ✅ **Message History** - yangi user ulanganda oxirgi 50 ta xabarni ko'rish
- ✅ **Join/Leave Notifications** - kim keldi/ketdi xabarlari
- ✅ **Input Validation** - username, room, message validation
- ✅ **Graceful Shutdown** - `context` bilan to'g'ri to'xtash
- ✅ **Thread-safe Operations** - `mutex` va `channel` bilan xavfsiz operatsiyalar

---

## Loyiha Strukturasi (Project Structure)

```
chatx/
├── cmd/server/main.go           # Entry point - server ishga tushirish
├── internal/
│   ├── config/config.go         # Konfiguratsiya (env variable'lar)
│   ├── handler/handler.go       # HTTP -> WebSocket upgrade
│   ├── message/message.go       # Message type'lari va struct'lar
│   ├── store/store.go           # In-memory xabar saqlash (history)
│   ├── validator/validator.go   # Input validation logikasi
│   └── websocket/
│       ├── hub.go               # Markaziy connection boshqaruvchi
│       └── client.go            # Har bir connection uchun handler
├── .env                         # Konfiguratsiya fayli (git'ga commit qilinmaydi)
├── .gitignore                   # Git ignore qoidalari
├── test-client.html             # Browser'da test qilish uchun client
└── claude.md                    # Bu fayl - Claude uchun context
```

### Har Bir Paket Nima Qiladi?

- **`cmd/server`** - Dasturning boshlanish nuqtasi. HTTP server yaratadi va Hub'ni ishga tushiradi.
- **`internal/config`** - Environment variable'larni o'qiydi va default qiymatlar beradi.
- **`internal/handler`** - HTTP request'ni WebSocket'ga upgrade qiladi va Client yaratadi.
- **`internal/message`** - Message struct'i va message type'lari (chat, dm, join, leave, typing, presence).
- **`internal/store`** - In-memory storage, har bir room uchun oxirgi xabarlarni saqlaydi.
- **`internal/validator`** - Username, room, message validation qiladi.
- **`internal/websocket/hub`** - Barcha Client'larni boshqaradi, xabarlarni tarqatadi (broadcast).
- **`internal/websocket/client`** - Bitta WebSocket connection uchun ReadPump va WritePump goroutine'lari.

---

## Arxitektura Pattern'lari

### 1. Hub-and-Spoke Pattern

Bu pattern'da **Hub** markazda turadi va barcha **Client**'lar unga ulangan (spoke = g'ildirak nurlari).

```
         [Browser 1] ← Client 1 ↘
         [Browser 2] ← Client 2 → Hub (goroutine)
         [Browser 3] ← Client 3 ↗
```

**Nima uchun bu pattern kerak?**
- Barcha Client'larni bir joydan boshqarish (centralized control)
- Thread-safety: faqat Hub `Clients` map'ini o'zgartiradi (race condition yo'q)
- Message broadcasting: bitta xabar kelsa, barchaga tarqatish oson

### 2. Concurrency Model (Goroutine + Channel)

**Hub:**
- Bitta goroutine'da ishlaydi
- `select` statement bilan 3 ta channel'ni tinglaydi:
  - `Register` channel - yangi Client ulanganda
  - `Unregister` channel - Client uzilganda
  - `Broadcast` channel - xabar tarqatish kerak bo'lganda

**Client (har biri uchun 2 ta goroutine):**
- `ReadPump()` goroutine - browser'dan xabar o'qiydi, Hub'ga yuboradi
- `WritePump()` goroutine - Hub'dan xabar olib, browser'ga yozadi

**Channel'lar nima uchun thread-safe?**
- Go'da channel'lar built-in lock mexanizmi bilan keladi
- Bir nechta goroutine bir vaqtda channel'ga yoza/o'qiy oladi
- Mutex qo'yish kerak emas

### 3. Thread-Safety Strategiyasi

| Resource          | Himoya Usuli                          | Sabab                                         |
|-------------------|---------------------------------------|-----------------------------------------------|
| `Hub.Clients`     | Hub goroutine (single writer)         | Faqat Hub o'zgartiradi, boshqalar faqat o'qiydi |
| `Hub.OnlineUsers` | `sync.Mutex` (Lock/Unlock)            | Bir nechta goroutine o'zgartirishi mumkin     |
| `Hub.UserClients` | `sync.Mutex` (Lock/Unlock)            | DM yuborishda concurrent access bo'ladi       |
| `Client.Send`     | Go channel (built-in thread-safe)     | Channel o'zi thread-safe                      |

---

## Message Type'lari

ChatX'da 6 xil message type bor:

| Type       | Tavsif                                      | History'da saqlanadimi? |
|------------|---------------------------------------------|-------------------------|
| `chat`     | Oddiy chat xabari (room'dagi barchaga)      | ✅ Ha                   |
| `dm`       | Direct Message (shaxsiy, 1-to-1)            | ✅ Ha                   |
| `join`     | User xonaga kirdi notification              | ✅ Ha                   |
| `leave`    | User xonadan chiqdi notification            | ✅ Ha                   |
| `typing`   | Typing indicator (kim yozyapti)             | ❌ Yo'q                 |
| `presence` | Online user'lar ro'yxati (JSON array)       | ❌ Yo'q                 |

---

## WebSocket API

### Connection URL Format
```
ws://localhost:8080/ws?username=Ali&room=general
```

**Query parameter'lar:**
- `username` - foydalanuvchi ismi (1-20 belgidan iborat)
- `room` - xona nomi (1-50 belgidan iborat)

### Message Format (JSON)
```json
{
  "type": "chat",
  "content": "Salom hammaga!",
  "username": "Ali",
  "room_id": "general",
  "timestamp": 1234567890,
  "recipient": ""  // Faqat DM type uchun
}
```

### Direct Message Qanday Yuborish?

Xabarni `@username` bilan boshlasangiz, avtomatik DM'ga aylanadi:

```
@john Salom, qalaysan?
```

Bu xabar:
1. Type `dm`'ga o'zgaradi
2. Faqat `john` va `Ali` (sender) ga yuboriladi
3. Boshqa hech kim ko'rmaydi

Kod ichida bu qanday ishlaydi? → `client.go:71-107` qismida:
```go
if strings.HasPrefix(msg.Content, "@") {
    parts := strings.SplitN(msg.Content, " ", 2)
    recipient := strings.TrimPrefix(parts[0], "@")
    msg.Type = message.MessageTypeDM
    msg.Recipient = recipient
    // ...
}
```

---

## Konfiguratsiya (Environment Variables)

`.env` faylda yoki terminal'da environment variable sifatida berish mumkin:

```bash
CHATX_PORT=:8080                    # Server porti
CHATX_STORE_MAX=100                 # Har bir room uchun max xabarlar soni
CHATX_HISTORY_COUNT=50              # Yangi user ulanganda nechta xabar yuborish
CHATX_MAX_MESSAGE_SIZE=512          # Xabar maksimal uzunligi (byte)
CHATX_WRITE_WAIT=10s                # Write timeout (WebSocket)
CHATX_PONG_WAIT=60s                 # Pong kutish vaqti (keep-alive)
CHATX_PING_PERIOD=54s               # Ping yuborish intervali
CHATX_LOG_LEVEL=info                # Log darajasi (info, debug, error)
```

**Terminal orqali o'zgartirish:**
```bash
CHATX_PORT=:8081 CHATX_STORE_MAX=200 go run cmd/server/main.go
```

---

## Development Commandlari

### Server'ni Ishga Tushirish
```bash
go run cmd/server/main.go
```

### Build Qilish
```bash
go build ./...
```

### Test Qilish
```bash
# test-client.html faylni browser'da oching
# Bir nechta tab ochib, turli username va room bilan ulaning
# DM test qiling: @username xabar
```

### Graceful Shutdown
```bash
Ctrl+C  # SIGINT signal yuboradi, server gracefully to'xtaydi
```

---

## Muhim Implementation Detallari

### 1. Nima uchun Channel Ishlatiladi?

**Channel'lar Go'da goroutine'lar o'rtasida xavfsiz aloqa qilish uchun ishlatiladi.**

```go
// ❌ Yomon - mutex bilan
type Hub struct {
    Clients map[*Client]bool
    mu      sync.Mutex
}

func (h *Hub) AddClient(c *Client) {
    h.mu.Lock()
    h.Clients[c] = true
    h.mu.Unlock()
}

// ✅ Yaxshi - channel bilan
type Hub struct {
    Register chan *Client
}

// Client'ni qo'shish kerak bo'lganda
hub.Register <- client  // Channel'ga yubor

// Hub goroutine'ida
case client := <-h.Register:
    h.Clients[client] = true  // Thread-safe, faqat Hub o'zgartiradi
```

**Afzalliklari:**
- Mutex qo'yish kerak emas (Go channel'ni o'zi lock qiladi)
- `select` statement bilan bir nechta channel'ni tinglash mumkin
- Deadlock'dan qochish oson

### 2. Nima uchun Hub Pattern?

**Agar Hub bo'lmasa:**
- Har bir Client boshqa barcha Client'larga to'g'ridan-to'g'ri xabar yuborishi kerak
- Bu N×N connection yaratadi (murakkab, xavfli)
- Race condition ko'p bo'ladi

**Hub pattern bilan:**
- Barcha Client faqat Hub'ga ulangan (N connection)
- Faqat Hub `Clients` map'ini o'zgartiradi (race condition yo'q)
- Yangi feature qo'shish oson (message filtering, logging, rate limiting)

### 3. Client Lifecycle (Hayot Tsikli)

```
1. Browser → HTTP Request → ws://localhost:8080/ws?username=Ali&room=general
                ↓
2. handler.ServeWS() → HTTP'ni WebSocket'ga upgrade qiladi
                ↓
3. Client struct yaratiladi → Hub.Register'ga yuboriladi
                ↓
4. Hub → client'ni Clients map'iga qo'shadi
                ↓
5. ReadPump() va WritePump() goroutine'lari ishga tushadi
                ↓
6. Client xabar yuboradi/oladi (ishlaydi)
                ↓
7. Connection uziladi / xato yuz beradi
                ↓
8. ReadPump defer → Hub.Unregister'ga yuboradi, connection yopiladi
                ↓
9. Hub → client'ni Clients map'idan o'chiradi, Send channel yopiladi
                ↓
10. WritePump → channel yopilganini biladi, to'xtaydi
```

### 4. Message Flow (Xabar Oqimi)

```
Browser (JSON xabar)
        ↓
WebSocket Connection
        ↓
ReadPump goroutine (client.go:33-117)
        ↓ (validation)
Hub.Broadcast channel'ga yuborish
        ↓
Hub.Run() goroutine (hub.go:163-184)
        ↓ (select case msg := <-h.Broadcast)
Store'ga saqlash + Barcha Client'larga yuborish
        ↓
Har bir Client.Send channel'ga yozish
        ↓
WritePump goroutine (client.go:132-183)
        ↓
WebSocket Connection
        ↓
Browser (JSON xabar oladi)
```

---

## Ma'lum Cheklovlar (Limitations)

- **In-memory storage** - server restart bo'lsa, xabarlar yo'qoladi
- **Autentifikatsiya yo'q** - hamma istalgan username bilan ulana oladi
- **Rate limiting yo'q** - spam qilish mumkin
- **Enkriptsiya yo'q** - xabarlar plain text (HTTPS emas)
- **Horizontal scaling yo'q** - bitta server, load balancer qo'shib bo'lmaydi

---

## Kelajakda Qo'shiladigan Funksiyalar (TODO)

- [ ] **Unit Test'lar** - `validator`, `store`, `hub` testlari
- [ ] **Rate Limiting** - spam prevention (masalan: 10 xabar/soniya max)
- [ ] **JWT Authentication** - token-based login
- [ ] **PostgreSQL** - xabarlarni database'da saqlash (persistence)
- [ ] **User Profile** - avatar, bio, status
- [ ] **File/Image Sharing** - rasm yuborish
- [ ] **Read Receipts** - kim o'qidi ko'rsatish
- [ ] **Message Search** - xabar qidirish
- [ ] **Docker** - containerization
- [ ] **Production Deploy** - AWS/GCP'ga deploy qilish

---

## Testing Strategiyasi

### Manual Testing
`test-client.html` fayldan foydalaning:

1. **Bir nechta room test:**
   - Browser'da 3 ta tab oching
   - 1-tab: `username=Ali, room=general`
   - 2-tab: `username=Vali, room=general`
   - 3-tab: `username=Sardor, room=tech`
   - Ali va Vali bir-birining xabarini ko'radi, Sardor ko'rmaydi

2. **DM test:**
   - Ali: `@Vali salom` yozsa
   - Faqat Ali va Vali ko'radi, boshqalar ko'rmaydi

3. **Connection drop test:**
   - Tab yoping, boshqa tab'larda "left the room" message ko'rinishi kerak

4. **Graceful shutdown test:**
   - Server'da `Ctrl+C` bosing
   - Log'da "graceful shutdown" ko'rinishi kerak

---

## Tez-tez Uchraydigan Muammolar va Yechimlar

### 1. Import Cycle (Circular Dependency)

**Muammo:** `package A` → `package B` → `package A` (infinite loop)

**Yechim:**
- Shared type'larni alohida package'ga ajrating (`internal/message`)
- `hub` va `client` bir package ichida bo'lishi kerak (`internal/websocket`)

### 2. Race Condition

**Muammo:** Bir nechta goroutine bir vaqtda map'ga yozmoqda

**Yechim:**
```go
// ❌ Yomon - race condition
h.OnlineUsers[roomID][username] = time.Now()

// ✅ Yaxshi - mutex bilan
h.mu.Lock()
h.OnlineUsers[roomID][username] = time.Now()
h.mu.Unlock()
```

**Race detector bilan test qiling:**
```bash
go run -race cmd/server/main.go
```

### 3. Connection Drop (Ping/Pong)

**Muammo:** Browser minimized qilganda yoki network uzilganda connection uziladi

**Yechim:**
- `WritePump` har 54 soniyada ping yuboradi (`client.go:146`)
- Browser avtomatik pong javobini yuboradi
- Agar pong kelmasa, connection "dead" deb hisoblanadi va yopiladi

---

## Code Style Va Best Practice'lar

### 1. Function'lar Kichik Bo'lsin
```go
// ❌ Yomon - 200 qatorli function
func (h *Hub) Run() {
    // ... juda ko'p kod ...
}

// ✅ Yaxshi - kichik function'lar
func (h *Hub) Run() {
    // ...
    h.handleRegister(client)
    h.handleBroadcast(msg)
}

func (h *Hub) handleRegister(client *Client) { /*...*/ }
func (h *Hub) handleBroadcast(msg *Message) { /*...*/ }
```

### 2. Descriptive Variable Name'lar
```go
// ❌ Yomon
c := &Client{...}
m := &Message{...}

// ✅ Yaxshi
client := &Client{...}
message := &Message{...}
```

### 3. Concurrency Logic'ni Comment Qiling
```go
// Nima uchun defer ishlatildi - cleanup uchun
// Agar connection uzilsa yoki panic bo'lsa, defer bajariladi
defer func() {
    h.Unregister <- c  // Hub'dan o'chirish
    c.Conn.Close()     // WebSocket yopish
}()
```

### 4. Error Handling - Panic Emas, Log
```go
// ❌ Yomon - panic (butun server crash bo'ladi)
if err != nil {
    panic(err)
}

// ✅ Yaxshi - log va davom etish
if err != nil {
    log.Printf("Xato: %v", err)
    return
}
```

### 5. Defer Cleanup Operatsiyalari Uchun
```go
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()    // Timer'ni to'xtatish
        c.Conn.Close()   // Connection yopish
    }()
    // ... kod ...
}
```

---

## Go'da O'rganilgan Konseptlar

Bu loyihada quyidagi Go konseptlari amaliyotda ishlatilgan:

### 1. Concurrency
- `goroutine` - lightweight thread
- `channel` - goroutine'lar o'rtasida aloqa
- `select` statement - bir nechta channel'ni tinglash
- `context` - graceful shutdown

### 2. Design Pattern'lar
- Hub-and-Spoke pattern
- Constructor pattern (`NewHub()`, `NewClient()`)
- Dependency Injection (`Handler`'ga `Hub` uzatish)

### 3. Thread-Safety
- `sync.Mutex` - shared resource himoyalash
- `sync.RWMutex` - read/write lock
- Channel - built-in thread-safe

### 4. Network Programming
- WebSocket protokoli
- HTTP handler'lar
- HTTP → WebSocket upgrade

### 5. Arxitektura
- Package separation (har bir package bitta vazifa)
- Import cycle'dan qochish
- `internal` package - private code

### 6. Configuration
- Environment variable'lar
- Default qiymatlar berish
- `.env` fayl o'qish

### 7. Error Handling
- Error check qilish (`if err != nil`)
- Log bilan xatolarni yozish
- Graceful degradation (crash emas, davom etish)

### 8. Data Structure'lar
- Map - key-value storage (`Hub.Clients`)
- Slice - dinamik array (`Store.Messages`)
- Channel - queue (FIFO)

---

## Claude Uchun Oxirgi Ko'rsatma

Agar foydalanuvchi savollar bersa yoki kod yozishni so'rasa:
1. **O'zbek tilida javob ber** (keywords ingliz tilida qolsin)
2. **Batafsil tushuntir** - nima, nima uchun, qanday
3. **Misol ko'rsat** - real code snippet bilan
4. **Concurrency pattern'lari bilan bog'la** - goroutine, channel, mutex konseptlarini eslatib o'tgin

Maqsad: Foydalanuvchi nafaqat kod yozishni, balki **Go backend'ning ichki ishlashini** chuqur tushunsin.
