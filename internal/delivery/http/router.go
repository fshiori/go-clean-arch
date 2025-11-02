package http

import (
	"go-clean-arch/internal/delivery/http/handler"

	"github.com/gin-gonic/gin"
)

// Router sets up all HTTP routes
func Router(
	userHandler *handler.UserHandler,
	// Add other handlers here as needed
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

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

		// Order routes can be added here
		// orders := v1.Group("/orders")
		// {
		//     orders.POST("", orderHandler.CreateOrder)
		//     orders.GET("/:id", orderHandler.GetOrder)
		//     ...
		// }
	}

	return router
}
