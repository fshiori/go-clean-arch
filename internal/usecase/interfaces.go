package usecase

import (
	"context"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
)

// UserUsecase defines the interface for user-related business logic
type UserUsecase interface {
	CreateUser(ctx context.Context, email, password string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUserPassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
}

// OrderUsecase defines the interface for order-related business logic
type OrderUsecase interface {
	CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error)
	Checkout(ctx context.Context, orderID int64, paymentInfo *port.PaymentInfo) (*port.Transaction, error)
	GetOrderByID(ctx context.Context, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error)
	ShipOrder(ctx context.Context, orderID int64) error
	CompleteOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) error
}
