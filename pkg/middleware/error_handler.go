// Package middleware provides HTTP middleware functions for the Gin framework.
package middleware

import (
	"net/http"
	"strings"

	"go-clean-arch/internal/domain"
	"go-clean-arch/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/samber/oops"
)

// ErrorResponse represents the standard error response format
type ErrorResponse struct {
	Error   string                 `json:"error"`             // Error message
	Code    string                 `json:"code,omitempty"`    // Error code
	TraceID string                 `json:"trace_id"`          // Trace ID for debugging
	Details map[string]interface{} `json:"details,omitempty"` // Additional details (dev mode only)
}

// HandleError provides unified error handling for HTTP responses
// It extracts information from oops errors and maps them to appropriate HTTP status codes
func HandleError(c *gin.Context, err error) {
	// Get trace ID from context
	traceID := GetTraceID(c)

	// Log the error
	logger.ErrorContext(c.Request.Context(), "Request failed",
		"error", err.Error(),
		"trace_id", traceID,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	// Default values
	var (
		statusCode = http.StatusInternalServerError
		errorCode  = "INTERNAL_ERROR"
		message    = "Internal server error"
		details    map[string]interface{}
	)

	// Check for Gin validation errors first
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
		message = "Request validation failed"
		details = make(map[string]interface{})
		for _, fieldErr := range validationErrs {
			details[fieldErr.Field()] = fieldErr.Tag()
		}
	} else if strings.Contains(err.Error(), "invalid character") ||
		strings.Contains(err.Error(), "unexpected end of JSON") ||
		strings.Contains(err.Error(), "cannot unmarshal") {
		// Handle JSON parsing errors
		statusCode = http.StatusBadRequest
		errorCode = "INVALID_JSON"
		message = err.Error()
	} else if oopsErr, ok := oops.AsOops(err); ok {
		// Try to extract oops error information
		// Extract error code
		if code := oopsErr.Code(); code != "" {
			errorCode = code

			// Map error code to HTTP status code using domain mapping
			if httpStatus, exists := domain.ErrorCodeToHTTPStatus[code]; exists {
				statusCode = httpStatus
			}
		}

		// Extract error message
		if msg := oopsErr.Error(); msg != "" {
			message = msg
		}

		// In debug mode, provide additional debugging information
		if gin.Mode() == gin.DebugMode {
			details = make(map[string]interface{})

			if domain := oopsErr.Domain(); domain != "" {
				details["domain"] = domain
			}

			if tags := oopsErr.Tags(); len(tags) > 0 {
				details["tags"] = tags
			}

			if ctx := oopsErr.Context(); len(ctx) > 0 {
				details["context"] = ctx
			}

			if hint := oopsErr.Hint(); hint != "" {
				details["hint"] = hint
			}

			if stacktrace := oopsErr.Stacktrace(); len(stacktrace) > 0 {
				details["stacktrace"] = stacktrace
			}
		}
	} else {
		// Standard Go error
		message = err.Error()
	}

	// Build error response
	response := ErrorResponse{
		Error:   message,
		Code:    errorCode,
		TraceID: traceID,
		Details: details,
	}

	// Return JSON response
	c.JSON(statusCode, response)
}
