package domain

import (
	"errors"
	"time"
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
	ID         int64
	UserID     int64
	TotalAmount float64
	Status     OrderStatus
	Items      []OrderItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID int64
	Quantity  int
	Price     float64
}

// NewOrder creates a new order
func NewOrder(userID int64, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	order := &Order{
		UserID:    userID,
		Status:    OrderStatusPending,
		Items:     items,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
		return errors.New("can only mark pending orders as paid")
	}
	o.Status = OrderStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

// Ship marks the order as shipped
func (o *Order) Ship() error {
	if o.Status != OrderStatusPaid {
		return errors.New("can only ship paid orders")
	}
	o.Status = OrderStatusShipped
	o.UpdatedAt = time.Now()
	return nil
}

// Complete marks the order as completed
func (o *Order) Complete() error {
	if o.Status != OrderStatusShipped {
		return errors.New("can only complete shipped orders")
	}
	o.Status = OrderStatusCompleted
	o.UpdatedAt = time.Now()
	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusCompleted || o.Status == OrderStatusCancelled {
		return errors.New("cannot cancel completed or already cancelled orders")
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
