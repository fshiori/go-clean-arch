package domain

import (
	"time"

	"github.com/samber/oops"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusShipped   OrderStatus = "SHIPPED"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// Order represents an order entity
type Order struct {
	ID          int64
	UserID      int64
	TotalAmount float64
	Status      OrderStatus
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID int64
	Quantity  int
	Price     float64
}

// NewOrder creates a new order
func NewOrder(userID int64, items []OrderItem) (*Order, error) {
	// Validate user ID
	if userID <= 0 {
		return nil, oops.
			Code(ErrCodeInvalidUserID).
			In("domain").
			With("user_id", userID).
			Hint("User ID must be a positive integer").
			Wrap(ErrInvalidUserID)
	}

	// Validate items
	if len(items) == 0 {
		return nil, oops.
			Code(ErrCodeEmptyOrder).
			In("domain").
			With("user_id", userID).
			Hint("Order must have at least one item").
			Wrap(ErrEmptyOrder)
	}

	// Validate each item
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, oops.
				Code(ErrCodeInvalidQuantity).
				In("domain").
				With("item_index", i).
				With("product_id", item.ProductID).
				With("quantity", item.Quantity).
				Hint("Item quantity must be positive").
				Wrap(ErrInvalidQuantity)
		}
		if item.Price < 0 {
			return nil, oops.
				Code(ErrCodeInvalidQuantity).
				In("domain").
				With("item_index", i).
				With("product_id", item.ProductID).
				With("price", item.Price).
				Hint("Item price cannot be negative").
				Wrap(ErrInvalidQuantity)
		}
	}

	now := time.Now()
	order := &Order{
		UserID:    userID,
		Status:    OrderStatusPending,
		Items:     items,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Calculate total amount
	order.calculateTotal()

	return order, nil
}

// calculateTotal calculates the total amount of the order
func (o *Order) calculateTotal() {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
}

// MarkAsPaid marks the order as paid
func (o *Order) MarkAsPaid() error {
	if o.Status != OrderStatusPending {
		return oops.
			Code(ErrCodeInvalidOrderStatus).
			In("domain").
			With("order_id", o.ID).
			With("current_status", o.Status).
			With("expected_status", OrderStatusPending).
			Hint("Can only mark pending orders as paid").
			Wrap(ErrInvalidOrderStatus)
	}
	o.Status = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

// Ship marks the order as shipped
func (o *Order) Ship() error {
	if o.Status != OrderStatusPaid {
		return oops.
			Code(ErrCodeInvalidOrderStatus).
			In("domain").
			With("order_id", o.ID).
			With("current_status", o.Status).
			With("expected_status", OrderStatusPaid).
			Hint("Can only ship paid orders").
			Wrap(ErrInvalidOrderStatus)
	}
	o.Status = OrderStatusShipped
	o.UpdatedAt = time.Now()
	return nil
}

// Complete marks the order as completed
func (o *Order) Complete() error {
	if o.Status != OrderStatusShipped {
		return oops.
			Code(ErrCodeInvalidOrderStatus).
			In("domain").
			With("order_id", o.ID).
			With("current_status", o.Status).
			With("expected_status", OrderStatusShipped).
			Hint("Can only complete shipped orders").
			Wrap(ErrInvalidOrderStatus)
	}
	o.Status = OrderStatusCompleted
	o.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusCompleted || o.Status == OrderStatusCancelled {
		return oops.
			Code(ErrCodeInvalidOrderStatus).
			In("domain").
			With("order_id", o.ID).
			With("current_status", o.Status).
			Hint("Cannot cancel completed or already cancelled orders").
			Wrap(ErrInvalidOrderStatus)
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
