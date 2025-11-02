package usecase

import (
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
)

// OrderInteractor handles order-related business logic
type OrderInteractor struct {
	orderRepo      port.OrderRepository
	userRepo       port.UserRepository
	paymentGateway port.PaymentGateway
}

// NewOrderInteractor creates a new OrderInteractor
func NewOrderInteractor(
	orderRepo port.OrderRepository,
	userRepo port.UserRepository,
	paymentGateway port.PaymentGateway,
) *OrderInteractor {
	return &OrderInteractor{
		orderRepo:      orderRepo,
		userRepo:       userRepo,
		paymentGateway: paymentGateway,
	}
}

// CreateOrder creates a new order
func (i *OrderInteractor) CreateOrder(userID int64, items []domain.OrderItem) (*domain.Order, error) {
	// Verify user exists
	_, err := i.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Create order using domain factory function
	order, err := domain.NewOrder(userID, items)
	if err != nil {
		return nil, err
	}

	// Save order
	if err := i.orderRepo.Save(order); err != nil {
		return nil, err
	}

	return order, nil
}

// Checkout processes payment for an order
func (i *OrderInteractor) Checkout(orderID int64, paymentInfo *port.PaymentInfo) (*port.Transaction, error) {
	// Retrieve order
	order, err := i.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	// Verify order is in pending status
	if order.Status != domain.OrderStatusPending {
		return nil, errors.New("can only checkout pending orders")
	}

	// Process payment through gateway (abstracted)
	transaction, err := i.paymentGateway.CreateTransaction(order, paymentInfo)
	if err != nil {
		return nil, err
	}

	// Check if payment was successful
	if transaction.Status == port.TransactionStatusCompleted {
		// Mark order as paid
		if err := order.MarkAsPaid(); err != nil {
			return nil, err
		}

		// Update order in repository
		if err := i.orderRepo.Update(order); err != nil {
			return nil, err
		}
	}

	return transaction, nil
}

// GetOrderByID retrieves an order by ID
func (i *OrderInteractor) GetOrderByID(orderID int64) (*domain.Order, error) {
	if orderID <= 0 {
		return nil, errors.New("invalid order ID")
	}

	order, err := i.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetUserOrders retrieves all orders for a user
func (i *OrderInteractor) GetUserOrders(userID int64) ([]*domain.Order, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	orders, err := i.orderRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// ShipOrder ships an order
func (i *OrderInteractor) ShipOrder(orderID int64) error {
	order, err := i.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	// Use domain method to ship (includes business rules validation)
	if err := order.Ship(); err != nil {
		return err
	}

	// Update in repository
	return i.orderRepo.Update(order)
}

// CompleteOrder marks an order as completed
func (i *OrderInteractor) CompleteOrder(orderID int64) error {
	order, err := i.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	// Use domain method to complete
	if err := order.Complete(); err != nil {
		return err
	}

	// Update in repository
	return i.orderRepo.Update(order)
}

// CancelOrder cancels an order
func (i *OrderInteractor) CancelOrder(orderID int64) error {
	order, err := i.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	// Use domain method to cancel
	if err := order.Cancel(); err != nil {
		return err
	}

	// Update in repository
	return i.orderRepo.Update(order)
}
