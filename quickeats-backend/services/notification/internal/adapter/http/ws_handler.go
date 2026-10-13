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

	claims, _ := token.Claims.(jwt.MapClaims)
	shopID, _ := claims["shop"].(string)
	if shopID == "" {
		conn.WriteJSON(map[string]string{"type": "ERROR", "message": "no shop in token"})
		conn.Close()
		return
	}

	// Register connection
	h.hub.Register(shopID, conn)
	conn.WriteJSON(map[string]string{"type": "CONNECTED", "shopId": shopID})

	// Keep connection alive, read messages (for future client-to-server messaging)
	defer h.hub.Unregister(shopID, conn)
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
