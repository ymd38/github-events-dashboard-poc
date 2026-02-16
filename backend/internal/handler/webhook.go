package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/service"
)

const signaturePrefix = "sha256="

// WebhookHandler handles POST /api/webhook requests from GitHub.
type WebhookHandler struct {
	webhookSecret string
	eventService  *service.EventService
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(webhookSecret string, eventService *service.EventService) *WebhookHandler {
	return &WebhookHandler{
		webhookSecret: webhookSecret,
		eventService:  eventService,
	}
}

// ServeHTTP handles the webhook request.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.LogEvent("error", "failed to read request body", map[string]interface{}{"error": err.Error()})
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()
	signature := r.Header.Get("X-Hub-Signature-256")
	if !h.verifySignature(body, signature) {
		middleware.LogEvent("warn", "webhook signature verification failed", map[string]interface{}{
			"remote_addr": r.RemoteAddr,
		})
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}
	deliveryID := r.Header.Get("X-GitHub-Delivery")
	eventType := r.Header.Get("X-GitHub-Event")
	if deliveryID == "" || eventType == "" {
		writeError(w, http.StatusBadRequest, "missing required headers")
		return
	}
	middleware.LogEvent("info", "webhook received", map[string]interface{}{
		"delivery_id": deliveryID,
		"event_type":  eventType,
	})
	result, savedEvent, err := h.eventService.ProcessWebhook(deliveryID, eventType, body)
	if err != nil {
		middleware.LogEvent("error", "failed to process webhook", map[string]interface{}{
			"delivery_id": deliveryID,
			"error":       err.Error(),
		})
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if savedEvent != nil {
		broadcastEvent(*savedEvent)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if !strings.HasPrefix(signature, signaturePrefix) {
		return false
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, signaturePrefix))
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(payload)
	expected := mac.Sum(nil)
	return hmac.Equal(sig, expected)
}

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Broadcast callback type for SSE integration (set later in Phase 5).
var BroadcastFunc func(event model.Event)

func broadcastEvent(event model.Event) {
	if BroadcastFunc != nil {
		BroadcastFunc(event)
	}
}
