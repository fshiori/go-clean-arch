package handler

import (
	"net/http"
	"strconv"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase"
	"go-clean-arch/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests related to orders
type OrderHandler struct {
	orderInteractor usecase.OrderUsecase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(orderInteractor usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{
		orderInteractor: orderInteractor,
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert DTO items to domain items
	items := make([]domain.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	// Call usecase
	order, err := h.orderInteractor.CreateOrder(c.Request.Context(), req.UserID, items)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entity to response DTO
	response := toOrderResponse(order)

	c.JSON(http.StatusCreated, response)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	// Call usecase
	order, err := h.orderInteractor.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entity to response DTO
	response := toOrderResponse(order)

	c.JSON(http.StatusOK, response)
}

// ListUserOrders handles GET /users/:userId/orders
func (h *OrderHandler) ListUserOrders(c *gin.Context) {
	// Parse user ID from URL
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Call usecase
	orders, err := h.orderInteractor.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entities to response DTOs
	response := toOrderListResponse(orders)

	c.JSON(http.StatusOK, response)
}

// Checkout handles POST /orders/:id/checkout
func (h *OrderHandler) Checkout(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert DTO to payment info
	paymentInfo := toPaymentInfo(&req)

	// Call usecase
	transaction, err := h.orderInteractor.Checkout(c.Request.Context(), id, paymentInfo)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert transaction to response
	response := toTransactionResponse(transaction)

	c.JSON(http.StatusOK, response)
}

// ShipOrder handles POST /orders/:id/ship
func (h *OrderHandler) ShipOrder(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	// Call usecase
	if err := h.orderInteractor.ShipOrder(c.Request.Context(), id); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order shipped successfully"})
}

// CompleteOrder handles POST /orders/:id/complete
func (h *OrderHandler) CompleteOrder(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	// Call usecase
	if err := h.orderInteractor.CompleteOrder(c.Request.Context(), id); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order completed successfully"})
}

// CancelOrder handles POST /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID format"})
		return
	}

	// Call usecase
	if err := h.orderInteractor.CancelOrder(c.Request.Context(), id); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled successfully"})
}
