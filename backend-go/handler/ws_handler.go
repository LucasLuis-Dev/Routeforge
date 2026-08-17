package handler

import (
	"net/http"

	ws "github.com/LucasLuis-Dev/Routeforge/backend-go/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Permite conexões CORS para testes de clientes web e móveis
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "falha no handshake do WebSocket", http.StatusBadRequest)
		return
	}

	client := &ws.Client{
		Hub:  h.hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()
}

// Método auxiliar de utilidade no Hub
func (h *WSHandler) GetHub() *ws.Hub {
	return h.hub
}
