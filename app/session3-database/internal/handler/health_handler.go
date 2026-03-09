package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
	db        *sql.DB
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		db:        db,
	}
}

// DatabaseHealth holds database connection pool information
type DatabaseHealth struct {
	Status          string `json:"status"`
	OpenConnections int    `json:"open_connections"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	MaxOpen         int    `json:"max_open"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string         `json:"status"`
	Database  DatabaseHealth `json:"database"`
	Timestamp time.Time      `json:"timestamp"`
}

// Check handles GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	dbHealth := DatabaseHealth{
		Status: "connected",
	}

	status := http.StatusOK
	overallStatus := "ok"

	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			dbHealth.Status = "disconnected"
			overallStatus = "degraded"
			status = http.StatusServiceUnavailable
		} else {
			stats := h.db.Stats()
			dbHealth.OpenConnections = stats.OpenConnections
			dbHealth.InUse = stats.InUse
			dbHealth.Idle = stats.Idle
			dbHealth.MaxOpen = stats.MaxOpenConnections
		}
	}

	response := HealthResponse{
		Status:    overallStatus,
		Database:  dbHealth,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response) //nolint:errcheck
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
