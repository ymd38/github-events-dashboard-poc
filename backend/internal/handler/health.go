package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check API response.
type HealthResponse struct {
	Status string       `json:"status"`
	Checks HealthChecks `json:"checks"`
}

// HealthChecks holds individual service check results.
type HealthChecks struct {
	Database string `json:"database"`
}

// HealthHandler handles GET /api/health requests.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// ServeHTTP handles the health check request.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dbStatus := "up"
	if err := h.db.Ping(); err != nil {
		dbStatus = "down"
	}
	overallStatus := "healthy"
	statusCode := http.StatusOK
	if dbStatus == "down" {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}
	resp := HealthResponse{
		Status: overallStatus,
		Checks: HealthChecks{Database: dbStatus},
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}
