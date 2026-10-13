package app

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/quickeats/services/notification/internal/domain"
)

// Hub maintains active WebSocket connections per shop.
type Hub struct {
	mu    sync.RWMutex
	shops map[string]map[*websocket.Conn]bool // shopID -> set of connections
}

func NewHub() *Hub {
	return &Hub{shops: make(map[string]map[*websocket.Conn]bool)}
}

func (h *Hub) Register(shopID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.shops[shopID] == nil {
		h.shops[shopID] = make(map[*websocket.Conn]bool)
	}
	h.shops[shopID][conn] = true
	slog.Info("ws: client registered", "shop_id", shopID)
}

func (h *Hub) Unregister(shopID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.shops[shopID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.shops, shopID)
		}
	}
	conn.Close()
	slog.Info("ws: client disconnected", "shop_id", shopID)
}

// SendToShop broadcasts a message to all connected clients for a given shop.
func (h *Hub) SendToShop(shopID string, msg domain.WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.shops[shopID]
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
			slog.Error("ws: write failed", "error", err)
			go h.Unregister(shopID, conn)
		}
	}
}

func (h *Hub) ActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, conns := range h.shops {
		count += len(conns)
	}
	return count
}
