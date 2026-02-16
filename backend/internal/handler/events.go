package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/service"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

// EventsHandler handles GET /api/events and GET /api/events/{id} requests.
type EventsHandler struct {
	eventService *service.EventService
}

// NewEventsHandler creates a new EventsHandler.
func NewEventsHandler(eventService *service.EventService) *EventsHandler {
	return &EventsHandler{eventService: eventService}
}

// List handles GET /api/events with pagination and optional event_type filter.
func (h *EventsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page := parseIntQuery(r, "page", defaultPage)
	perPage := parseIntQuery(r, "per_page", defaultPerPage)
	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 || perPage > maxPerPage {
		perPage = defaultPerPage
	}
	eventType := r.URL.Query().Get("event_type")
	result, err := h.eventService.ListEvents(page, perPage, eventType)
	if err != nil {
		middleware.LogEvent("error", "failed to list events", map[string]interface{}{"error": err.Error()})
		writeError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// GetByID handles GET /api/events/{id}.
func (h *EventsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	event, err := h.eventService.GetEventByID(id)
	if err != nil {
		middleware.LogEvent("error", "failed to get event", map[string]interface{}{"error": err.Error(), "id": id})
		writeError(w, http.StatusInternalServerError, "failed to get event")
		return
	}
	if event == nil {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(event)
}

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
