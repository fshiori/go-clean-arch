package port

import "go-clean-arch/internal/domain"

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	// FindByID retrieves an order by its ID
	FindByID(id int64) (*domain.Order, error)

	// FindByUserID retrieves all orders for a specific user
	FindByUserID(userID int64) ([]*domain.Order, error)

	// Save creates a new order
	Save(order *domain.Order) error

	// Update updates an existing order
	Update(order *domain.Order) error

	// UpdateStatus updates the status of an order
	UpdateStatus(orderID int64, status domain.OrderStatus) error

	// List retrieves orders with pagination
	List(offset, limit int) ([]*domain.Order, error)
}
