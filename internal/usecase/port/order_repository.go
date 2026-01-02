// Package port defines the interfaces (ports) required by the use case layer.
package port

import (
	"context"
	"go-clean-arch/internal/domain"
)

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	// FindByID retrieves an order by its ID
	FindByID(ctx context.Context, id int64) (*domain.Order, error)

	// FindByUserID retrieves all orders for a specific user
	FindByUserID(ctx context.Context, userID int64) ([]*domain.Order, error)

	// Save creates a new order
	Save(ctx context.Context, order *domain.Order) error

	// Update updates an existing order
	Update(ctx context.Context, order *domain.Order) error

	// UpdateStatus updates the status of an order
	UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error

	// List retrieves orders with pagination
	List(ctx context.Context, offset, limit int) ([]*domain.Order, error)
}
