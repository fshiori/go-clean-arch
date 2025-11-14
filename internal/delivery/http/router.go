package http

import (
	"go-clean-arch/internal/delivery/http/handler"
	"go-clean-arch/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// Router sets up all HTTP routes
func Router(
	userHandler *handler.UserHandler,
	orderHandler *handler.OrderHandler,
	healthHandler *handler.HealthHandler,
) *gin.Engine {
	// Create router without default middleware
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add trace ID middleware (must be before logger)
	router.Use(middleware.TraceID())

	// Add request logger middleware
	router.Use(middleware.RequestLogger())

	// Health check endpoints (for Kubernetes probes and monitoring)
	router.GET("/health", healthHandler.HealthCheck)      // Comprehensive health check
	router.GET("/ready", healthHandler.ReadinessCheck)    // Readiness probe
	router.GET("/live", healthHandler.LivenessCheck)      // Liveness probe

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id/password", userHandler.UpdatePassword)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.POST("/:id/checkout", orderHandler.Checkout)
			orders.POST("/:id/ship", orderHandler.ShipOrder)
			orders.POST("/:id/complete", orderHandler.CompleteOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// User's orders
		v1.GET("/users/:userId/orders", orderHandler.ListUserOrders)
	}

	return router
}
