package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	Uptime    time.Duration `json:"uptime_seconds"`
	Timestamp time.Time     `json:"timestamp"`
}

// Check handles GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Message:   "Mini ASM service is running",
		Uptime:    time.Since(h.startTime),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

/*
🎓 NOTES:

Refactored từ Buổi 1:
- Buổi 1: Health check logic trong main.go
- Buổi 2: Extracted to separate handler

Benefits:
- Consistent with other handlers
- Can add more health checks (database, etc.) in Buổi 3
- Reusable and testable
*/
