# edu-tga Chat Application

## Architecture Overview
This is a **real-time chat application** built with pure Go (minimal third-party dependencies). The architecture follows a **hub-and-spoke pattern** for WebSocket connections:

- **Hub** (`internal/websocket/hub.go`): Central message dispatcher that manages all active client connections. Runs in its own goroutine with channels for register/unregister/broadcast operations.
- **Client** (`internal/websocket/client.go`): Represents each WebSocket connection. Each client runs two goroutines: `readPump()` (reads from browser → hub) and `writePump()` (writes from hub → browser).
- **Message** (`internal/websocket/message.go`): JSON-serialized message structure with types: chat, join, leave, typing.

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
cmd/server/main.go           # Entry point (empty, being developed)
internal/
  websocket/
    hub.go                   # Connection manager
    client.go                # Per-connection handler
    message.go               # Message types
  handler/                   # HTTP → WebSocket upgrade (empty)
  store/                     # In-memory storage (empty)
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
**In Progress:** User is learning to build this step-by-step
- ✅ Message structures defined
- ✅ Hub with broadcast/register/unregister logic
- ✅ Client with readPump/writePump methods
- ⏳ HTTP handlers for WebSocket upgrade
- ⏳ Main server setup
- ⏳ In-memory store

## Important Notes
- This is a **learning project** - user is writing code themselves with guidance
- Prioritize educational clarity over production patterns
- Keep explanations in Uzbek when requested
- Third-party libraries are intentionally avoided to understand Go fundamentals
