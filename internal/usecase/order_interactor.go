package usecase

import (
	"context"
	"errors"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/samber/oops"
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
func (i *OrderInteractor) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	// Verify user exists
	_, err := i.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("order_usecase").
				Tags("order_creation", "validation").
				With("user_id", userID).
				Hint("User must exist before creating an order").
				Wrap(err)
		}

		return nil, oops.
			Code(domain.ErrCodeUserFetchFailed).
			In("order_usecase").
			Tags("order_creation").
			With("user_id", userID).
			Wrapf(err, "failed to verify user existence")
	}

	// Create order using domain factory function
	order, err := domain.NewOrder(userID, items)
	if err != nil {
		return nil, oops.
			In("order_usecase").
			Tags("order_creation", "validation").
			With("user_id", userID).
			With("item_count", len(items)).
			Wrapf(err, "failed to create order entity")
	}

	// Save order
	if err := i.orderRepo.Save(ctx, order); err != nil {
		return nil, oops.
			Code(domain.ErrCodeOrderSaveFailed).
			In("order_usecase").
			Tags("order_creation", "database").
			With("user_id", userID).
			Hint("Check database connection and constraints").
			Wrapf(err, "failed to save order to repository")
	}

	return order, nil
}

// Checkout processes payment for an order
func (i *OrderInteractor) Checkout(ctx context.Context, orderID int64, paymentInfo *port.PaymentInfo) (*port.Transaction, error) {
	// Retrieve order
	order, err := i.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, oops.
				Code(domain.ErrCodeOrderNotFound).
				In("order_usecase").
				Tags("checkout").
				With("order_id", orderID).
				Wrap(err)
		}

		return nil, oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			Tags("checkout").
			With("order_id", orderID).
			Wrapf(err, "failed to fetch order")
	}

	// Verify order is in pending status
	if order.Status != domain.OrderStatusPending {
		return nil, oops.
			Code(domain.ErrCodeInvalidOrderStatus).
			In("order_usecase").
			Tags("checkout", "validation").
			With("order_id", orderID).
			With("current_status", order.Status).
			With("expected_status", domain.OrderStatusPending).
			Hint("Can only checkout pending orders").
			Wrap(domain.ErrInvalidOrderStatus)
	}

	// Process payment through gateway (abstracted)
	transaction, err := i.paymentGateway.CreateTransaction(order, paymentInfo)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodePaymentFailed).
			In("order_usecase").
			Tags("checkout", "payment").
			With("order_id", orderID).
			With("amount", order.TotalAmount).
			Hint("Check payment gateway connection and payment info").
			Wrapf(err, "payment gateway transaction failed")
	}

	// Check if payment was successful
	if transaction.Status == port.TransactionStatusCompleted {
		// Mark order as paid
		if err := order.MarkAsPaid(); err != nil {
			return nil, oops.
				In("order_usecase").
				Tags("checkout", "status_update").
				With("order_id", orderID).
				Wrapf(err, "failed to mark order as paid")
		}

		// Update order in repository
		if err := i.orderRepo.Update(ctx, order); err != nil {
			return nil, oops.
				Code(domain.ErrCodeOrderUpdateFailed).
				In("order_usecase").
				Tags("checkout", "database").
				With("order_id", orderID).
				Hint("Order was paid but database update failed - manual intervention may be required").
				Wrapf(err, "failed to update order after payment")
		}
	}

	return transaction, nil
}

// GetOrderByID retrieves an order by ID
func (i *OrderInteractor) GetOrderByID(ctx context.Context, orderID int64) (*domain.Order, error) {
	if orderID <= 0 {
		return nil, oops.
			Code(domain.ErrCodeInvalidUserID).
			In("order_usecase").
			With("order_id", orderID).
			Hint("Order ID must be a positive integer").
			Wrap(errors.New("invalid order ID"))
	}

	order, err := i.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, oops.
				Code(domain.ErrCodeOrderNotFound).
				In("order_usecase").
				With("order_id", orderID).
				Wrap(err)
		}

		return nil, oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			With("order_id", orderID).
			Hint("Check database connection").
			Wrapf(err, "failed to fetch order")
	}

	return order, nil
}

// GetUserOrders retrieves all orders for a user
func (i *OrderInteractor) GetUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	if userID <= 0 {
		return nil, oops.
			Code(domain.ErrCodeInvalidUserID).
			In("order_usecase").
			With("user_id", userID).
			Wrap(domain.ErrInvalidUserID)
	}

	orders, err := i.orderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			Tags("listing").
			With("user_id", userID).
			Hint("Check database connection").
			Wrapf(err, "failed to fetch user orders")
	}

	return orders, nil
}

// ShipOrder ships an order
func (i *OrderInteractor) ShipOrder(ctx context.Context, orderID int64) error {
	order, err := i.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return oops.
				Code(domain.ErrCodeOrderNotFound).
				In("order_usecase").
				Tags("shipping").
				With("order_id", orderID).
				Wrap(err)
		}

		return oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			Tags("shipping").
			With("order_id", orderID).
			Wrapf(err, "failed to fetch order")
	}

	// Use domain method to ship (includes business rules validation)
	if err := order.Ship(); err != nil {
		return oops.
			In("order_usecase").
			Tags("shipping", "validation").
			With("order_id", orderID).
			Wrapf(err, "failed to ship order")
	}

	// Update in repository
	if err := i.orderRepo.Update(ctx, order); err != nil {
		return oops.
			Code(domain.ErrCodeOrderUpdateFailed).
			In("order_usecase").
			Tags("shipping", "database").
			With("order_id", orderID).
			Wrapf(err, "failed to update order after shipping")
	}

	return nil
}

// CompleteOrder marks an order as completed
func (i *OrderInteractor) CompleteOrder(ctx context.Context, orderID int64) error {
	order, err := i.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return oops.
				Code(domain.ErrCodeOrderNotFound).
				In("order_usecase").
				Tags("completion").
				With("order_id", orderID).
				Wrap(err)
		}

		return oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			Tags("completion").
			With("order_id", orderID).
			Wrapf(err, "failed to fetch order")
	}

	// Use domain method to complete
	if err := order.Complete(); err != nil {
		return oops.
			In("order_usecase").
			Tags("completion", "validation").
			With("order_id", orderID).
			Wrapf(err, "failed to complete order")
	}

	// Update in repository
	if err := i.orderRepo.Update(ctx, order); err != nil {
		return oops.
			Code(domain.ErrCodeOrderUpdateFailed).
			In("order_usecase").
			Tags("completion", "database").
			With("order_id", orderID).
			Wrapf(err, "failed to update order after completion")
	}

	return nil
}

// CancelOrder cancels an order
func (i *OrderInteractor) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := i.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return oops.
				Code(domain.ErrCodeOrderNotFound).
				In("order_usecase").
				Tags("cancellation").
				With("order_id", orderID).
				Wrap(err)
		}

		return oops.
			Code(domain.ErrCodeOrderFetchFailed).
			In("order_usecase").
			Tags("cancellation").
			With("order_id", orderID).
			Wrapf(err, "failed to fetch order")
	}

	// Use domain method to cancel
	if err := order.Cancel(); err != nil {
		return oops.
			In("order_usecase").
			Tags("cancellation", "validation").
			With("order_id", orderID).
			Wrapf(err, "failed to cancel order")
	}

	// Update in repository
	if err := i.orderRepo.Update(ctx, order); err != nil {
		return oops.
			Code(domain.ErrCodeOrderUpdateFailed).
			In("order_usecase").
			Tags("cancellation", "database").
			With("order_id", orderID).
			Wrapf(err, "failed to update order after cancellation")
	}

	return nil
}
