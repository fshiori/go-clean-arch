package usecase

import (
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
)

// UserUsecase defines the interface for user-related business logic
type UserUsecase interface {
	CreateUser(email, password string) (*domain.User, error)
	GetUserByID(id int64) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUserPassword(userID int64, oldPassword, newPassword string) error
	ListUsers(page, pageSize int) ([]*domain.User, error)
	DeleteUser(id int64) error
}

// OrderUsecase defines the interface for order-related business logic
type OrderUsecase interface {
	CreateOrder(userID int64, items []domain.OrderItem) (*domain.Order, error)
	Checkout(orderID int64, paymentInfo *port.PaymentInfo) (*port.Transaction, error)
	GetOrderByID(orderID int64) (*domain.Order, error)
	GetUserOrders(userID int64) ([]*domain.Order, error)
	ShipOrder(orderID int64) error
	CompleteOrder(orderID int64) error
	CancelOrder(orderID int64) error
}
