package handler

import (
	"fmt"
	"net/http"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/sse"
)

// SSEHandler handles GET /api/events/stream for Server-Sent Events.
type SSEHandler struct {
	hub *sse.Hub
}

// NewSSEHandler creates a new SSEHandler.
func NewSSEHandler(hub *sse.Hub) *SSEHandler {
	return &SSEHandler{hub: hub}
}

// ServeHTTP streams events to the client via SSE.
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	client := make(chan []byte, 64)
	h.hub.Register(client)
	defer h.hub.Unregister(client)
	ctx := r.Context()
	middleware.LogEvent("info", "SSE stream started", map[string]interface{}{
		"remote_addr": r.RemoteAddr,
	})
	for {
		select {
		case <-ctx.Done():
			middleware.LogEvent("info", "SSE stream closed by client", map[string]interface{}{
				"remote_addr": r.RemoteAddr,
			})
			return
		case data, ok := <-client:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: new_event\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
