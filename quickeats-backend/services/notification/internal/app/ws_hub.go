package app

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/quickeats/services/notification/internal/domain"
)

// Room keys: a shop's staff devices share one room, each customer has their own.
func ShopRoom(shopID string) string     { return "shop:" + shopID }
func CustomerRoom(userID string) string { return "customer:" + userID }

// Hub maintains active WebSocket connections per room.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[*websocket.Conn]bool // room -> set of connections
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]bool)}
}

func (h *Hub) Register(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*websocket.Conn]bool)
	}
	h.rooms[room][conn] = true
	slog.Info("ws: client registered", "room", room)
}

func (h *Hub) Unregister(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.rooms[room]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.rooms, room)
		}
	}
	conn.Close()
	slog.Info("ws: client disconnected", "room", room)
}

// SendToShop broadcasts a message to all connected devices of a shop.
func (h *Hub) SendToShop(shopID string, msg domain.WSMessage) {
	h.send(ShopRoom(shopID), msg)
}

// SendToCustomer sends a message to all connected devices of a customer.
func (h *Hub) SendToCustomer(userID string, msg domain.WSMessage) {
	h.send(CustomerRoom(userID), msg)
}

// send holds the full lock while writing: gorilla/websocket allows only one
// concurrent writer per connection, and events arrive on several consumer goroutines.
func (h *Hub) send(room string, msg domain.WSMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.rooms[room]
	if !ok {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("ws: marshal failed", "error", err)
		return
	}

	for conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			slog.Error("ws: write failed", "room", room, "error", err)
			go h.Unregister(room, conn)
		}
	}
}

func (h *Hub) ActiveConnections() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for _, conns := range h.rooms {
		count += len(conns)
	}
	return count
}
