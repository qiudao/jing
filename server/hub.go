package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) Send(conn *websocket.Conn, msgType string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	msg := Message{Type: msgType, Data: raw}
	payload, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, payload)
}
