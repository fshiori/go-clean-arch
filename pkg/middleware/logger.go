// Package middleware provides HTTP middleware functions for the Gin framework.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"go-clean-arch/pkg/logger"
)

// RequestLogger logs HTTP requests with structured logging.
// It captures comprehensive request and response information including:
//   - Request method, path, and query parameters
//   - Response status code
//   - Request processing latency
//   - Client IP address and User-Agent
//   - Trace ID for request correlation
//   - Any errors that occurred during request processing
//
// The middleware should be registered after the TraceID middleware
// to ensure trace IDs are available for logging.
//
// Usage:
//
//	router.Use(middleware.TraceID())
//	router.Use(middleware.RequestLogger())
//
// Log output example (JSON format):
//
//	{
//	  "time": "2025-11-02T10:30:45Z",
//	  "level": "INFO",
//	  "msg": "HTTP request completed",
//	  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
//	  "method": "POST",
//	  "path": "/api/v1/users",
//	  "status": 201,
//	  "latency_ms": 45,
//	  "client_ip": "192.168.1.100"
//	}
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get trace ID
		traceID := GetTraceID(c)

		// Log request details with structured fields
		logger.InfoContext(c.Request.Context(),
			"HTTP request completed",
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Log errors if any occurred during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.ErrorContext(c.Request.Context(),
					"Request error",
					"trace_id", traceID,
					"error", err.Error(),
				)
			}
		}
	}
}
