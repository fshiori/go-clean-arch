package handler

import (
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
)

// CreateOrderRequest is the DTO for creating an order
type CreateOrderRequest struct {
	UserID int64            `json:"user_id" binding:"required,gt=0"`
	Items  []OrderItemDTO   `json:"items" binding:"required,min=1,dive"`
}

// OrderItemDTO represents an order item in the API layer
type OrderItemDTO struct {
	ProductID int64   `json:"product_id" binding:"required,gt=0"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gte=0"`
}

// CheckoutRequest is the DTO for checkout
type CheckoutRequest struct {
	CardNumber string `json:"card_number" binding:"required"`
	CVV        string `json:"cvv" binding:"required,len=3"`
	ExpiryDate string `json:"expiry_date" binding:"required"`
}

// OrderResponse is the DTO for order response
type OrderResponse struct {
	ID          int64            `json:"id"`
	UserID      int64            `json:"user_id"`
	TotalAmount float64          `json:"total_amount"`
	Status      string           `json:"status"`
	Items       []OrderItemDTO   `json:"items"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// OrderListResponse is the DTO for list of orders
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int             `json:"total_count"`
}

// TransactionResponse is the DTO for transaction response
type TransactionResponse struct {
	ID              string                 `json:"id"`
	OrderID         int64                  `json:"order_id"`
	Amount          float64                `json:"amount"`
	Status          string                 `json:"status"`
	PaymentMethod   string                 `json:"payment_method"`
	TransactionData map[string]interface{} `json:"transaction_data,omitempty"`
}

// toOrderResponse converts a domain.Order to OrderResponse DTO
func toOrderResponse(order *domain.Order) OrderResponse {
	items := make([]OrderItemDTO, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemDTO{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return OrderResponse{
		ID:          order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		Items:       items,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

// toOrderListResponse converts a list of domain.Order to OrderListResponse DTO
func toOrderListResponse(orders []*domain.Order) OrderListResponse {
	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = toOrderResponse(order)
	}

	return OrderListResponse{
		Orders:     orderResponses,
		TotalCount: len(orders),
	}
}

// toPaymentInfo converts CheckoutRequest to port.PaymentInfo
func toPaymentInfo(req *CheckoutRequest) *port.PaymentInfo {
	return &port.PaymentInfo{
		CardNumber: req.CardNumber,
		CVV:        req.CVV,
		ExpiryDate: req.ExpiryDate,
	}
}

// toTransactionResponse converts port.Transaction to TransactionResponse DTO
func toTransactionResponse(tx *port.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:              tx.ID,
		OrderID:         tx.OrderID,
		Amount:          tx.Amount,
		Status:          string(tx.Status),
		PaymentMethod:   tx.PaymentMethod,
		TransactionData: tx.TransactionData,
	}
}
