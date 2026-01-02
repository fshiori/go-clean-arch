package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *sqlx.DB
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *sqlx.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       string            `json:"status"`
	Timestamp    time.Time         `json:"timestamp"`
	Dependencies map[string]string `json:"dependencies"`
}

// HealthCheck handles GET /health
// Returns the health status of the service and its dependencies
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dependencies := make(map[string]string)
	overallStatus := "healthy"

	// Check database connection
	dbStatus := h.checkDatabase(ctx)
	dependencies["database"] = dbStatus
	if dbStatus != "healthy" {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:       overallStatus,
		Timestamp:    time.Now(),
		Dependencies: dependencies,
	}

	statusCode := http.StatusOK
	if overallStatus == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// checkDatabase verifies the database connection
func (h *HealthHandler) checkDatabase(ctx context.Context) string {
	if h.db == nil {
		return "unavailable"
	}

	// Ping the database
	err := h.db.PingContext(ctx)
	if err != nil {
		return "unhealthy"
	}

	return "healthy"
}

// ReadinessCheck handles GET /ready
// Returns whether the service is ready to accept traffic
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check if database is ready
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":  false,
			"reason": "database not initialized",
		})
		return
	}

	err := h.db.PingContext(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready":  false,
			"reason": "database not ready",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}

// LivenessCheck handles GET /live
// Returns whether the service is alive (for Kubernetes liveness probes)
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"alive": true,
	})
}
