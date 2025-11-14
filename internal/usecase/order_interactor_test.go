package usecase

import (
	"context"
	"testing"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
	"go-clean-arch/internal/usecase/port/mocks"

	"github.com/samber/oops"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type OrderInteractorTestSuite struct {
	suite.Suite
	orderRepo      *mocks.MockOrderRepository
	userRepo       *mocks.MockUserRepository
	paymentGateway *mocks.MockPaymentGateway
	interactor     *OrderInteractor
	ctx            context.Context
}

func TestOrderInteractorTestSuite(t *testing.T) {
	suite.Run(t, new(OrderInteractorTestSuite))
}

func (s *OrderInteractorTestSuite) SetupTest() {
	s.orderRepo = new(mocks.MockOrderRepository)
	s.userRepo = new(mocks.MockUserRepository)
	s.paymentGateway = new(mocks.MockPaymentGateway)
	s.interactor = NewOrderInteractor(s.orderRepo, s.userRepo, s.paymentGateway)
	s.ctx = context.Background()
}

func (s *OrderInteractorTestSuite) TearDownTest() {
	s.orderRepo.AssertExpectations(s.T())
	s.userRepo.AssertExpectations(s.T())
	s.paymentGateway.AssertExpectations(s.T())
}

// Helper to create valid order items
func (s *OrderInteractorTestSuite) createValidItems() []domain.OrderItem {
	return []domain.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 1000},
		{ProductID: 2, Quantity: 1, Price: 2000},
	}
}

// Test CreateOrder
func (s *OrderInteractorTestSuite) TestCreateOrder_Success() {
	userID := int64(123)
	items := s.createValidItems()

	// Mock: User exists
	user := domain.ReconstructUser(userID, "user@example.com", "hashedpass", time.Now(), time.Now())
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Mock: Save succeeds
	s.orderRepo.On("Save", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.UserID == userID && len(o.Items) == 2
	})).Run(func(args mock.Arguments) {
		order := args.Get(1).(*domain.Order)
		order.ID = 456
	}).Return(nil)

	// Execute
	order, err := s.interactor.CreateOrder(s.ctx, userID, items)

	// Assert
	s.NoError(err)
	s.NotNil(order)
	s.Equal(userID, order.UserID)
	s.Equal(int64(456), order.ID)
	s.Equal(float64(4000), order.TotalAmount)
	s.Equal(domain.OrderStatusPending, order.Status)
}

func (s *OrderInteractorTestSuite) TestCreateOrder_UserNotFound() {
	userID := int64(999)
	items := s.createValidItems()

	// Mock: User doesn't exist
	s.userRepo.On("FindByID", s.ctx, userID).Return(nil, domain.ErrUserNotFound)

	// Execute
	order, err := s.interactor.CreateOrder(s.ctx, userID, items)

	// Assert
	s.Error(err)
	s.Nil(order)
	s.ErrorIs(err, domain.ErrUserNotFound)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(domain.ErrCodeUserNotFound, oopsErr.Code())
}

func (s *OrderInteractorTestSuite) TestCreateOrder_InvalidUserID() {
	items := s.createValidItems()

	// Mock: User repo returns error
	s.userRepo.On("FindByID", s.ctx, int64(0)).Return(nil, domain.ErrInvalidUserID)

	// Execute
	order, err := s.interactor.CreateOrder(s.ctx, 0, items)

	// Assert
	s.Error(err)
	s.Nil(order)
}

func (s *OrderInteractorTestSuite) TestCreateOrder_EmptyItems() {
	userID := int64(123)

	// Mock: User exists
	user := domain.ReconstructUser(userID, "user@example.com", "hashedpass", time.Now(), time.Now())
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Execute
	order, err := s.interactor.CreateOrder(s.ctx, userID, []domain.OrderItem{})

	// Assert
	s.Error(err)
	s.Nil(order)
	s.ErrorIs(err, domain.ErrEmptyOrder)
}

func (s *OrderInteractorTestSuite) TestCreateOrder_RepositoryError() {
	userID := int64(123)
	items := s.createValidItems()

	user := domain.ReconstructUser(userID, "user@example.com", "hashedpass", time.Now(), time.Now())
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Mock: Save fails
	dbErr := oops.Code("DB_ERROR").Errorf("database connection failed")
	s.orderRepo.On("Save", s.ctx, mock.AnythingOfType("*domain.Order")).Return(dbErr)

	// Execute
	order, err := s.interactor.CreateOrder(s.ctx, userID, items)

	// Assert
	s.Error(err)
	s.Nil(order)
}

// Test Checkout
func (s *OrderInteractorTestSuite) TestCheckout_Success() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID

	paymentInfo := &port.PaymentInfo{
		CardNumber: "4111111111111111",
		CVV:        "123",
		ExpiryDate: "12/25",
	}

	transaction := &port.Transaction{
		ID:            "txn_123",
		OrderID:       orderID,
		Amount:        order.TotalAmount,
		Status:        port.TransactionStatusCompleted,
		PaymentMethod: "credit_card",
	}

	// Mock: Order exists and is pending
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Mock: Payment succeeds
	s.paymentGateway.On("CreateTransaction", order, paymentInfo).Return(transaction, nil)

	// Mock: Order update succeeds
	s.orderRepo.On("Update", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusPaid
	})).Return(nil)

	// Execute
	txn, err := s.interactor.Checkout(s.ctx, orderID, paymentInfo)

	// Assert
	s.NoError(err)
	s.NotNil(txn)
	s.Equal("txn_123", txn.ID)
	s.Equal(port.TransactionStatusCompleted, txn.Status)
	s.Equal(domain.OrderStatusPaid, order.Status)
}

func (s *OrderInteractorTestSuite) TestCheckout_OrderNotFound() {
	orderID := int64(999)
	paymentInfo := &port.PaymentInfo{}

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(nil, domain.ErrOrderNotFound)

	// Execute
	txn, err := s.interactor.Checkout(s.ctx, orderID, paymentInfo)

	// Assert
	s.Error(err)
	s.Nil(txn)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

func (s *OrderInteractorTestSuite) TestCheckout_OrderNotPending() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid() // Already paid

	paymentInfo := &port.PaymentInfo{}

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Execute
	txn, err := s.interactor.Checkout(s.ctx, orderID, paymentInfo)

	// Assert
	s.Error(err)
	s.Nil(txn)
	s.ErrorIs(err, domain.ErrInvalidOrderStatus)
}

func (s *OrderInteractorTestSuite) TestCheckout_PaymentFailed() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID

	paymentInfo := &port.PaymentInfo{}

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Mock: Payment fails
	paymentErr := oops.Code("PAYMENT_DECLINED").Errorf("insufficient funds")
	s.paymentGateway.On("CreateTransaction", order, paymentInfo).Return(nil, paymentErr)

	// Execute
	txn, err := s.interactor.Checkout(s.ctx, orderID, paymentInfo)

	// Assert
	s.Error(err)
	s.Nil(txn)
}

// Test GetOrderByID
func (s *OrderInteractorTestSuite) TestGetOrderByID_Success() {
	orderID := int64(456)
	expectedOrder, _ := domain.NewOrder(123, s.createValidItems())
	expectedOrder.ID = orderID

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(expectedOrder, nil)

	// Execute
	order, err := s.interactor.GetOrderByID(s.ctx, orderID)

	// Assert
	s.NoError(err)
	s.NotNil(order)
	s.Equal(orderID, order.ID)
}

func (s *OrderInteractorTestSuite) TestGetOrderByID_NotFound() {
	orderID := int64(999)

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(nil, domain.ErrOrderNotFound)

	// Execute
	order, err := s.interactor.GetOrderByID(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.Nil(order)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

func (s *OrderInteractorTestSuite) TestGetOrderByID_InvalidID() {
	invalidIDs := []int64{0, -1, -100}

	for _, id := range invalidIDs {
		// Execute
		order, err := s.interactor.GetOrderByID(s.ctx, id)

		// Assert
		s.Error(err, "Should fail for ID: %d", id)
		s.Nil(order)
	}
}

// Test GetUserOrders
func (s *OrderInteractorTestSuite) TestGetUserOrders_Success() {
	userID := int64(123)

	order1, _ := domain.NewOrder(userID, s.createValidItems())
	order1.ID = 1
	order2, _ := domain.NewOrder(userID, s.createValidItems())
	order2.ID = 2

	expectedOrders := []*domain.Order{order1, order2}

	// Mock
	s.orderRepo.On("FindByUserID", s.ctx, userID).Return(expectedOrders, nil)

	// Execute
	orders, err := s.interactor.GetUserOrders(s.ctx, userID)

	// Assert
	s.NoError(err)
	s.Len(orders, 2)
	s.Equal(int64(1), orders[0].ID)
	s.Equal(int64(2), orders[1].ID)
}

func (s *OrderInteractorTestSuite) TestGetUserOrders_InvalidUserID() {
	// Execute
	orders, err := s.interactor.GetUserOrders(s.ctx, 0)

	// Assert
	s.Error(err)
	s.Nil(orders)
	s.ErrorIs(err, domain.ErrInvalidUserID)
}

func (s *OrderInteractorTestSuite) TestGetUserOrders_EmptyResult() {
	userID := int64(123)

	// Mock
	s.orderRepo.On("FindByUserID", s.ctx, userID).Return([]*domain.Order{}, nil)

	// Execute
	orders, err := s.interactor.GetUserOrders(s.ctx, userID)

	// Assert
	s.NoError(err)
	s.Empty(orders)
}

// Test ShipOrder
func (s *OrderInteractorTestSuite) TestShipOrder_Success() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid()

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)
	s.orderRepo.On("Update", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusShipped
	})).Return(nil)

	// Execute
	err := s.interactor.ShipOrder(s.ctx, orderID)

	// Assert
	s.NoError(err)
	s.Equal(domain.OrderStatusShipped, order.Status)
}

func (s *OrderInteractorTestSuite) TestShipOrder_NotPaid() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	// Order is still pending

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Execute
	err := s.interactor.ShipOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrInvalidOrderStatus)
}

func (s *OrderInteractorTestSuite) TestShipOrder_OrderNotFound() {
	orderID := int64(999)

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(nil, domain.ErrOrderNotFound)

	// Execute
	err := s.interactor.ShipOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

// Test CompleteOrder
func (s *OrderInteractorTestSuite) TestCompleteOrder_Success() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid()
	order.Ship()

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)
	s.orderRepo.On("Update", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusCompleted
	})).Return(nil)

	// Execute
	err := s.interactor.CompleteOrder(s.ctx, orderID)

	// Assert
	s.NoError(err)
	s.Equal(domain.OrderStatusCompleted, order.Status)
}

func (s *OrderInteractorTestSuite) TestCompleteOrder_NotShipped() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid()
	// Not shipped yet

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Execute
	err := s.interactor.CompleteOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrInvalidOrderStatus)
}

func (s *OrderInteractorTestSuite) TestCompleteOrder_OrderNotFound() {
	orderID := int64(999)

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(nil, domain.ErrOrderNotFound)

	// Execute
	err := s.interactor.CompleteOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

// Test CancelOrder
func (s *OrderInteractorTestSuite) TestCancelOrder_FromPending() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)
	s.orderRepo.On("Update", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusCancelled
	})).Return(nil)

	// Execute
	err := s.interactor.CancelOrder(s.ctx, orderID)

	// Assert
	s.NoError(err)
	s.Equal(domain.OrderStatusCancelled, order.Status)
}

func (s *OrderInteractorTestSuite) TestCancelOrder_FromPaid() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid()

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)
	s.orderRepo.On("Update", s.ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.ID == orderID && o.Status == domain.OrderStatusCancelled
	})).Return(nil)

	// Execute
	err := s.interactor.CancelOrder(s.ctx, orderID)

	// Assert
	s.NoError(err)
	s.Equal(domain.OrderStatusCancelled, order.Status)
}

func (s *OrderInteractorTestSuite) TestCancelOrder_AlreadyCompleted() {
	orderID := int64(456)
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = orderID
	order.MarkAsPaid()
	order.Ship()
	order.Complete()

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(order, nil)

	// Execute
	err := s.interactor.CancelOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrInvalidOrderStatus)
}

func (s *OrderInteractorTestSuite) TestCancelOrder_OrderNotFound() {
	orderID := int64(999)

	// Mock
	s.orderRepo.On("FindByID", s.ctx, orderID).Return(nil, domain.ErrOrderNotFound)

	// Execute
	err := s.interactor.CancelOrder(s.ctx, orderID)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}
