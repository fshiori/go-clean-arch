package domain

import (
	"testing"
	"time"

	"github.com/samber/oops"
	"github.com/stretchr/testify/suite"
)

type OrderTestSuite struct {
	suite.Suite
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}

// Helper to create valid order items
func (s *OrderTestSuite) createValidItems() []OrderItem {
	return []OrderItem{
		{ProductID: 1, Quantity: 2, Price: 1000},
		{ProductID: 2, Quantity: 1, Price: 2000},
	}
}

// Test NewOrder function
func (s *OrderTestSuite) TestNewOrder_Success() {
	userID := int64(123)
	items := s.createValidItems()

	order, err := NewOrder(userID, items)

	s.NoError(err)
	s.NotNil(order)
	s.Equal(userID, order.UserID)
	s.Equal(OrderStatusPending, order.Status)
	s.Equal(float64(4000), order.TotalAmount) // 2*1000 + 1*2000
	s.Len(order.Items, 2)
	s.NotZero(order.CreatedAt)
	s.NotZero(order.UpdatedAt)
}

func (s *OrderTestSuite) TestNewOrder_InvalidUserID() {
	invalidUserIDs := []int64{0, -1, -100}

	for _, userID := range invalidUserIDs {
		_, err := NewOrder(userID, s.createValidItems())

		s.Error(err, "Should fail for userID: %d", userID)
		s.ErrorIs(err, ErrInvalidUserID)

		oopsErr, ok := oops.AsOops(err)
		s.True(ok)
		s.Equal(ErrCodeInvalidUserID, oopsErr.Code())
	}
}

func (s *OrderTestSuite) TestNewOrder_EmptyItems() {
	_, err := NewOrder(123, []OrderItem{})

	s.Error(err)
	s.ErrorIs(err, ErrEmptyOrder)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(ErrCodeEmptyOrder, oopsErr.Code())
}

func (s *OrderTestSuite) TestNewOrder_NilItems() {
	_, err := NewOrder(123, nil)

	s.Error(err)
	s.ErrorIs(err, ErrEmptyOrder)
}

func (s *OrderTestSuite) TestNewOrder_InvalidItemQuantity() {
	items := []OrderItem{
		{ProductID: 1, Quantity: 0, Price: 1000},
	}

	_, err := NewOrder(123, items)

	s.Error(err)
	s.ErrorIs(err, ErrInvalidQuantity)
}

func (s *OrderTestSuite) TestNewOrder_InvalidItemPrice() {
	items := []OrderItem{
		{ProductID: 1, Quantity: 1, Price: -100},
	}

	_, err := NewOrder(123, items)

	s.Error(err)
	s.ErrorIs(err, ErrInvalidQuantity)
}

func (s *OrderTestSuite) TestNewOrder_InvalidProductID() {
	// Note: Current implementation doesn't validate ProductID
	// This test documents that behavior
	items := []OrderItem{
		{ProductID: 0, Quantity: 1, Price: 1000},
	}

	order, err := NewOrder(123, items)

	// Currently passes - ProductID validation not implemented
	s.NoError(err)
	s.NotNil(order)
}

func (s *OrderTestSuite) TestNewOrder_CalculateTotalAmount() {
	items := []OrderItem{
		{ProductID: 1, Quantity: 3, Price: 500},
		{ProductID: 2, Quantity: 2, Price: 1000},
		{ProductID: 3, Quantity: 1, Price: 2500},
	}

	order, err := NewOrder(123, items)

	s.NoError(err)
	expectedTotal := float64(3*500 + 2*1000 + 1*2500) // 6000
	s.Equal(expectedTotal, order.TotalAmount)
}

// Test MarkAsPaid
func (s *OrderTestSuite) TestMarkAsPaid_Success() {
	order, _ := NewOrder(123, s.createValidItems())

	err := order.MarkAsPaid()

	s.NoError(err)
	s.Equal(OrderStatusPaid, order.Status)
}

func (s *OrderTestSuite) TestMarkAsPaid_AlreadyPaid() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()

	err := order.MarkAsPaid()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

func (s *OrderTestSuite) TestMarkAsPaid_FromShipped() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()
	order.Ship()

	err := order.MarkAsPaid()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

// Test Ship
func (s *OrderTestSuite) TestShip_Success() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()

	err := order.Ship()

	s.NoError(err)
	s.Equal(OrderStatusShipped, order.Status)
}

func (s *OrderTestSuite) TestShip_NotPaid() {
	order, _ := NewOrder(123, s.createValidItems())

	err := order.Ship()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(ErrCodeInvalidOrderStatus, oopsErr.Code())
	s.Contains(oopsErr.Hint(), "paid")
}

func (s *OrderTestSuite) TestShip_AlreadyShipped() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()
	order.Ship()

	err := order.Ship()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

// Test Complete
func (s *OrderTestSuite) TestComplete_Success() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()
	order.Ship()

	err := order.Complete()

	s.NoError(err)
	s.Equal(OrderStatusCompleted, order.Status)
}

func (s *OrderTestSuite) TestComplete_NotShipped() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()

	err := order.Complete()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Contains(oopsErr.Hint(), "shipped")
}

func (s *OrderTestSuite) TestComplete_FromPending() {
	order, _ := NewOrder(123, s.createValidItems())

	err := order.Complete()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

// Test Cancel
func (s *OrderTestSuite) TestCancel_FromPending() {
	order, _ := NewOrder(123, s.createValidItems())

	err := order.Cancel()

	s.NoError(err)
	s.Equal(OrderStatusCancelled, order.Status)
}

func (s *OrderTestSuite) TestCancel_FromPaid() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()

	err := order.Cancel()

	s.NoError(err)
	s.Equal(OrderStatusCancelled, order.Status)
}

func (s *OrderTestSuite) TestCancel_FromShipped() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()
	order.Ship()

	err := order.Cancel()

	// Note: Current implementation allows cancelling shipped orders
	// This might be changed in the future based on business requirements
	s.NoError(err)
	s.Equal(OrderStatusCancelled, order.Status)
}

func (s *OrderTestSuite) TestCancel_FromCompleted() {
	order, _ := NewOrder(123, s.createValidItems())
	order.MarkAsPaid()
	order.Ship()
	order.Complete()

	err := order.Cancel()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

func (s *OrderTestSuite) TestCancel_AlreadyCancelled() {
	order, _ := NewOrder(123, s.createValidItems())
	order.Cancel()

	err := order.Cancel()

	s.Error(err)
	s.ErrorIs(err, ErrInvalidOrderStatus)
}

// Test Order state transitions
func (s *OrderTestSuite) TestOrderLifecycle_HappyPath() {
	order, err := NewOrder(123, s.createValidItems())
	s.NoError(err)
	s.Equal(OrderStatusPending, order.Status)

	// Pending -> Paid
	err = order.MarkAsPaid()
	s.NoError(err)
	s.Equal(OrderStatusPaid, order.Status)

	// Paid -> Shipped
	err = order.Ship()
	s.NoError(err)
	s.Equal(OrderStatusShipped, order.Status)

	// Shipped -> Completed
	err = order.Complete()
	s.NoError(err)
	s.Equal(OrderStatusCompleted, order.Status)
}

func (s *OrderTestSuite) TestOrderLifecycle_CancellationPath() {
	order, err := NewOrder(123, s.createValidItems())
	s.NoError(err)

	// Pending -> Paid
	err = order.MarkAsPaid()
	s.NoError(err)

	// Paid -> Cancelled (allowed)
	err = order.Cancel()
	s.NoError(err)
	s.Equal(OrderStatusCancelled, order.Status)
}

// Test edge cases
func (s *OrderTestSuite) TestOrder_LargeQuantity() {
	items := []OrderItem{
		{ProductID: 1, Quantity: 1000000, Price: 100},
	}

	order, err := NewOrder(123, items)

	s.NoError(err)
	s.Equal(float64(100000000), order.TotalAmount)
}

func (s *OrderTestSuite) TestOrder_MultipleItemsSameProduct() {
	// Business logic allows multiple items with same product ID
	// (This might be changed in the future to consolidate)
	items := []OrderItem{
		{ProductID: 1, Quantity: 2, Price: 1000},
		{ProductID: 1, Quantity: 3, Price: 1000},
	}

	order, err := NewOrder(123, items)

	s.NoError(err)
	s.Len(order.Items, 2)
	s.Equal(float64(5000), order.TotalAmount) // 2*1000 + 3*1000
}

func (s *OrderTestSuite) TestOrder_TimestampConsistency() {
	order, err := NewOrder(123, s.createValidItems())
	s.NoError(err)

	s.WithinDuration(order.CreatedAt, order.UpdatedAt, time.Second)
}
