package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/quickeats/services/notification/internal/app"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub       *app.Hub
	jwtSecret string
}

func NewWSHandler(hub *app.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "error", err)
		return
	}

	// Wait for AUTH message
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(msg, &authMsg); err != nil || !strings.EqualFold(authMsg.Type, "AUTH") {
		conn.WriteJSON(map[string]string{"type": "ERROR", "message": "expected AUTH message"})
		conn.Close()
		return
	}

	// Validate JWT
	token, err := jwt.Parse(authMsg.Token, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		conn.WriteJSON(map[string]string{"type": "ERROR", "message": "invalid token"})
		conn.Close()
		return
	}

	// Shop staff join their shop's room; customers join their own room.
	claims, _ := token.Claims.(jwt.MapClaims)
	shopID, _ := claims["shop"].(string)
	userID, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)

	var room string
	switch {
	case shopID != "":
		room = app.ShopRoom(shopID)
	case role == "customer" && userID != "":
		room = app.CustomerRoom(userID)
	default:
		conn.WriteJSON(map[string]string{"type": "ERROR", "message": "token has no shop or customer"})
		conn.Close()
		return
	}

	// Send the confirmation before registering, so it never races with an event write.
	conn.WriteJSON(map[string]string{"type": "CONNECTED", "room": room})
	h.hub.Register(room, conn)

	// Keep connection alive, read messages (for future client-to-server messaging)
	defer h.hub.Unregister(room, conn)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
