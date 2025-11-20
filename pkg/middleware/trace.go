// Package middleware provides HTTP middleware functions for the Gin framework.
// This package includes trace ID generation and request logging middleware.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-clean-arch/pkg/logger"
)

const (
	// TraceIDHeader is the HTTP header name for trace ID.
	// Clients can send this header to specify their own trace ID,
	// otherwise a new UUID will be generated.
	TraceIDHeader = "X-Trace-ID"

	// TraceIDKey is the context key for storing trace ID in gin.Context.
	// Use GetTraceID() to retrieve this value safely.
	TraceIDKey = "trace_id"
)

// TraceID middleware adds trace ID to request context for distributed tracing.
// It performs the following operations:
//   1. Checks if X-Trace-ID header exists in the request
//   2. If present, uses it; otherwise generates a new UUID
//   3. Stores the trace ID in gin.Context for easy access
//   4. Adds X-Trace-ID to response headers
//   5. Attaches trace ID to request context for structured logging
//
// Usage:
//   router.Use(middleware.TraceID())
//
// This middleware should be registered before any logging middleware
// to ensure all logs include the trace ID.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get trace ID from header or generate new one
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Store in gin context
		c.Set(TraceIDKey, traceID)

		// Add to response header
		c.Header(TraceIDHeader, traceID)

		// Create context with trace ID for logging
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetTraceID retrieves the trace ID from gin.Context.
// Returns empty string if trace ID is not found.
//
// Usage:
//   traceID := middleware.GetTraceID(c)
//   if traceID != "" {
//       // Use trace ID
//   }
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
