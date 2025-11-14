
## Edu-TGA Chat App - Go WebSocket Chat Server
- WebSocket Real-time Chat - bir vaqtda xabar almashish
- Hub-and-Spoke Pattern - markazlashtirilgan connection management
- Goroutines & Channels - concurrent programming
- Thread-safe Storage - Mutex bilan xavfsiz xotira
- Message History - yangi user eski xabarlarni ko'radi
- Multiple Rooms - turli xonalar, alohida history
- Join/Leave Messages - kim keldi/ketdi xabarlari

## Go'da o'rganganlar
- Concurrency: goroutines, channels, select statement
- Patterns: Hub pattern, Constructor pattern
- Thread-safety: sync.RWMutex, Lock/Unlock
- Network: WebSocket, HTTP handlers
- Architecture: Package separation, import cycles hal qilish
- Data structures: Maps, slices, channels

## Keyingi qadamlar
- User List - kimlar online ko'rsatish
- Typing Indicator - kim yozyapti
- Private Messages - shaxsiy xabarlar
- Authentication - login/password
- Database - PostgreSQL yoki MongoDB
- Deploy - production'ga chiqarish