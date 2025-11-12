package websocket

type Hub struct {
	// Registered clients
	clients map[*Client]bool
	// Inbound messages from the clients
	broadcast chan *Message
	// Register requests from the clients
	register chan *Client
	// Unregister requests from clients
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message muvaffaqiyatli yuborildi
				default:
					// Mijoz javob bermayotganda uni ro'yxatdan o'chirish
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
