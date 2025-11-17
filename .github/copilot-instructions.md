# ChatX - Real-time WebSocket Chat Application

## Architecture Overview
This is a **real-time chat application** built with pure Go (minimal third-party dependencies). The architecture follows a **hub-and-spoke pattern** for WebSocket connections:

- **Hub** (`internal/websocket/hub.go`): Central message dispatcher that manages all active client connections. Runs in its own goroutine with channels for register/unregister/broadcast operations.
- **Client** (`internal/websocket/client.go`): Represents each WebSocket connection. Each client runs two goroutines: `readPump()` (reads from browser → hub) and `writePump()` (writes from hub → browser).
- **Message** (`internal/message/message.go`): JSON-serialized message structure with types: chat, join, leave, typing.

### Data Flow
```
Browser → WebSocket → Client.readPump() → Hub.broadcast → Client.writePump() → WebSocket → Browser(s)
```

## Key Conventions

### Language & Communication
- **Codebase uses Uzbek language** for comments and discussions (e.g., "Mijoz javob bermayotganda", "Ping yuborish")
- Variable names and function signatures remain in English
- This is an **educational project** - code is meant for learning Go concurrency patterns

### WebSocket Pattern
- Uses `golang.org/x/net/websocket` (NOT gorilla/websocket) - only external dependency
- All WebSocket communication is JSON-encoded via `websocket.JSON.Send/Receive`
- Connection health maintained via ping/pong with 60s pong timeout, 54s ping period
- Max message size: 512 bytes

### Concurrency Model
- Hub runs in a single goroutine with `select` statement listening to 3 channels
- Each client spawns exactly 2 goroutines (read + write)
- Channel-based communication ensures thread-safety (no mutexes in client code)
- Client cleanup happens via `defer` in pump methods

### Project Structure
```
cmd/server/main.go           # Entry point with graceful shutdown
internal/
  config/
    config.go                # Environment-based configuration
  handler/
    handler.go               # HTTP → WebSocket upgrade handler
  message/
    message.go               # Message types and constants
  store/
    store.go                 # In-memory storage with RWMutex
  validator/
    validator.go             # Input validation (username, room, message)
  websocket/
    hub.go                   # Connection manager
    client.go                # Per-connection handler
test-client.html             # Browser test client
.env (optional)              # Configuration file
```

## Development Workflow

### Setup
```bash
go mod download
go run cmd/server/main.go
```

### No External Services
- No database - uses in-memory storage (maps with mutex for thread-safety)
- No message queues - Go channels handle all messaging
- No external auth - basic username-based identification

## Implementation Status
**✅ COMPLETED:** All core features are implemented and working
- ✅ Message structures defined (chat, join, leave types)
- ✅ Hub with broadcast/register/unregister logic
- ✅ Client with readPump/writePump methods
- ✅ HTTP handlers for WebSocket upgrade
- ✅ Main server setup with graceful shutdown (context-based)
- ✅ In-memory store with thread-safe operations (RWMutex)
- ✅ Configuration management (environment variables)
- ✅ Input validation (username, room, message content)
- ✅ Room-based messaging (multiple chat rooms)
- ✅ Message history (last 50 messages per room)
- ✅ Join/Leave notifications
- ✅ Browser test client (test-client.html)

## Key Features
- **Multiple Rooms:** Each room has isolated message history
- **Message History:** New users see last 50 messages when joining
- **Thread-Safe Storage:** RWMutex for concurrent read/write operations
- **Graceful Shutdown:** Context propagation from main → hub → clients
- **Configuration:** ENV variables with sensible defaults (CHATX_PORT, CHATX_STORE_MAX, etc.)
- **Validation:** Username (1-20 chars), Room (max 30 chars), Message (1-500 chars)

## Important Notes
- This is a **completed educational project** - demonstrates Go concurrency patterns
- Code includes extensive Uzbek comments for learning purposes
- Prioritize educational clarity over production patterns
- Third-party libraries are intentionally minimal to understand Go fundamentals
- No compilation errors, ready to run with `go run cmd/server/main.go`
