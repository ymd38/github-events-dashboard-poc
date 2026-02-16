package sse

import (
	"encoding/json"
	"sync"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
)

// Hub manages SSE client connections and broadcasts events to all connected clients.
type Hub struct {
	clients    map[chan []byte]bool
	register   chan chan []byte
	unregister chan chan []byte
	broadcast  chan model.Event
	mu         sync.RWMutex
}

// NewHub creates a new SSE Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[chan []byte]bool),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		broadcast:  make(chan model.Event, 256),
	}
}

// Run starts the hub event loop. Should be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			middleware.LogEvent("info", "SSE client connected", map[string]interface{}{
				"total_clients": h.ClientCount(),
			})
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
			h.mu.Unlock()
			middleware.LogEvent("info", "SSE client disconnected", map[string]interface{}{
				"total_clients": h.ClientCount(),
			})
		case event := <-h.broadcast:
			data, err := json.Marshal(event)
			if err != nil {
				middleware.LogEvent("error", "failed to marshal SSE event", map[string]interface{}{
					"error": err.Error(),
				})
				continue
			}
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client <- data:
				default:
					go func(c chan []byte) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register adds a new client channel to the hub.
func (h *Hub) Register(client chan []byte) {
	h.register <- client
}

// Unregister removes a client channel from the hub.
func (h *Hub) Unregister(client chan []byte) {
	h.unregister <- client
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(event model.Event) {
	h.broadcast <- event
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
