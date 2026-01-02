package repository

import (
	"context"
	"testing"

	"go-clean-arch/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type OrderRepositorySQLXTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *orderRepositorySQLX
}

func TestOrderRepositorySQLXTestSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositorySQLXTestSuite))
}

func (s *OrderRepositorySQLXTestSuite) SetupSuite() {
	// Create in-memory SQLite database
	db, err := sqlx.Connect("sqlite3", ":memory:")
	s.Require().NoError(err)

	// Create orders table
	schema := `
	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		total_amount REAL NOT NULL,
		status TEXT NOT NULL,
		items TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	s.Require().NoError(err)

	s.db = db
	s.repo = &orderRepositorySQLX{db: db}
}

func (s *OrderRepositorySQLXTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *OrderRepositorySQLXTestSuite) TearDownTest() {
	// Clean up orders table after each test
	_, err := s.db.Exec("DELETE FROM orders")
	s.Require().NoError(err)
}

// Helper to create valid order items
func (s *OrderRepositorySQLXTestSuite) createValidItems() []domain.OrderItem {
	return []domain.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 1000},
		{ProductID: 2, Quantity: 1, Price: 2000},
	}
}

// Test Save
func (s *OrderRepositorySQLXTestSuite) TestSave_Success() {
	ctx := context.Background()
	order, err := domain.NewOrder(123, s.createValidItems())
	s.Require().NoError(err)

	// Save order
	err = s.repo.Save(ctx, order)
	s.NoError(err)
	s.NotZero(order.ID)
	s.NotZero(order.CreatedAt)
	s.NotZero(order.UpdatedAt)
	s.Equal(float64(4000), order.TotalAmount)
	s.Equal(domain.OrderStatusPending, order.Status)
}

func (s *OrderRepositorySQLXTestSuite) TestSave_MultipleOrders() {
	ctx := context.Background()

	// Create and save multiple orders
	order1, _ := domain.NewOrder(100, s.createValidItems())
	err := s.repo.Save(ctx, order1)
	s.NoError(err)

	order2, _ := domain.NewOrder(200, s.createValidItems())
	err = s.repo.Save(ctx, order2)
	s.NoError(err)

	// Verify both have unique IDs
	s.NotZero(order1.ID)
	s.NotZero(order2.ID)
	s.NotEqual(order1.ID, order2.ID)
}

// Test FindByID
func (s *OrderRepositorySQLXTestSuite) TestFindByID_Success() {
	ctx := context.Background()
	order, _ := domain.NewOrder(123, s.createValidItems())

	// Save order first
	err := s.repo.Save(ctx, order)
	s.Require().NoError(err)

	// Find by ID
	found, err := s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(order.ID, found.ID)
	s.Equal(order.UserID, found.UserID)
	s.Equal(order.TotalAmount, found.TotalAmount)
	s.Equal(order.Status, found.Status)
	s.Len(found.Items, 2)
}

func (s *OrderRepositorySQLXTestSuite) TestFindByID_NotFound() {
	ctx := context.Background()

	// Try to find non-existent order
	found, err := s.repo.FindByID(ctx, 99999)
	s.Error(err)
	s.Nil(found)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

// Test FindByUserID
func (s *OrderRepositorySQLXTestSuite) TestFindByUserID_Success() {
	ctx := context.Background()
	userID := int64(123)

	// Create multiple orders for the same user
	order1, _ := domain.NewOrder(userID, s.createValidItems())
	err := s.repo.Save(ctx, order1)
	s.Require().NoError(err)

	order2, _ := domain.NewOrder(userID, s.createValidItems())
	err = s.repo.Save(ctx, order2)
	s.Require().NoError(err)

	// Create order for different user
	order3, _ := domain.NewOrder(456, s.createValidItems())
	err = s.repo.Save(ctx, order3)
	s.Require().NoError(err)

	// Find by user ID
	found, err := s.repo.FindByUserID(ctx, userID)
	s.NoError(err)
	s.Len(found, 2)
	for _, order := range found {
		s.Equal(userID, order.UserID)
	}
}

func (s *OrderRepositorySQLXTestSuite) TestFindByUserID_NotFound() {
	ctx := context.Background()

	// Try to find orders for non-existent user
	found, err := s.repo.FindByUserID(ctx, 99999)
	s.NoError(err)
	s.Empty(found)
}

// Test Update
func (s *OrderRepositorySQLXTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	order, _ := domain.NewOrder(123, s.createValidItems())

	// Save order first
	err := s.repo.Save(ctx, order)
	s.Require().NoError(err)

	originalUpdatedAt := order.UpdatedAt

	// Mark order as paid
	err = order.MarkAsPaid()
	s.Require().NoError(err)

	// Update order
	err = s.repo.Update(ctx, order)
	s.NoError(err)
	s.True(order.UpdatedAt.After(originalUpdatedAt))

	// Verify update
	found, err := s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Equal(domain.OrderStatusPaid, found.Status)
}

func (s *OrderRepositorySQLXTestSuite) TestUpdate_NotFound() {
	ctx := context.Background()
	order, _ := domain.NewOrder(123, s.createValidItems())
	order.ID = 99999 // Non-existent ID

	// Try to update non-existent order
	err := s.repo.Update(ctx, order)
	s.Error(err)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

// Test UpdateStatus
func (s *OrderRepositorySQLXTestSuite) TestUpdateStatus_Success() {
	ctx := context.Background()
	order, _ := domain.NewOrder(123, s.createValidItems())

	// Save order first
	err := s.repo.Save(ctx, order)
	s.Require().NoError(err)

	// Update status
	err = s.repo.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid)
	s.NoError(err)

	// Verify update
	found, err := s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Equal(domain.OrderStatusPaid, found.Status)
}

func (s *OrderRepositorySQLXTestSuite) TestUpdateStatus_NotFound() {
	ctx := context.Background()

	// Try to update status of non-existent order
	err := s.repo.UpdateStatus(ctx, 99999, domain.OrderStatusPaid)
	s.Error(err)
	s.ErrorIs(err, domain.ErrOrderNotFound)
}

// Test List
func (s *OrderRepositorySQLXTestSuite) TestList_Success() {
	ctx := context.Background()

	// Create multiple orders
	for i := 1; i <= 5; i++ {
		order, _ := domain.NewOrder(int64(100+i), s.createValidItems())
		err := s.repo.Save(ctx, order)
		s.Require().NoError(err)
	}

	// List orders with pagination
	orders, err := s.repo.List(ctx, 0, 3)
	s.NoError(err)
	s.Len(orders, 3)

	// List next page
	orders, err = s.repo.List(ctx, 3, 3)
	s.NoError(err)
	s.Len(orders, 2)
}

func (s *OrderRepositorySQLXTestSuite) TestList_Empty() {
	ctx := context.Background()

	// List from empty table
	orders, err := s.repo.List(ctx, 0, 10)
	s.NoError(err)
	s.Empty(orders)
}

// Test Order Items JSON serialization
func (s *OrderRepositorySQLXTestSuite) TestSave_ComplexOrderItems() {
	ctx := context.Background()

	// Create order with multiple items
	items := []domain.OrderItem{
		{ProductID: 1, Quantity: 2, Price: 1000},
		{ProductID: 2, Quantity: 1, Price: 2000},
		{ProductID: 3, Quantity: 5, Price: 500},
	}
	order, _ := domain.NewOrder(123, items)

	// Save order
	err := s.repo.Save(ctx, order)
	s.Require().NoError(err)

	// Retrieve order
	found, err := s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Len(found.Items, 3)

	// Verify items are correctly serialized/deserialized
	s.Equal(int64(1), found.Items[0].ProductID)
	s.Equal(2, found.Items[0].Quantity)
	s.Equal(float64(1000), found.Items[0].Price)

	s.Equal(int64(2), found.Items[1].ProductID)
	s.Equal(1, found.Items[1].Quantity)
	s.Equal(float64(2000), found.Items[1].Price)

	s.Equal(int64(3), found.Items[2].ProductID)
	s.Equal(5, found.Items[2].Quantity)
	s.Equal(float64(500), found.Items[2].Price)
}

// Test Order lifecycle
func (s *OrderRepositorySQLXTestSuite) TestOrderLifecycle() {
	ctx := context.Background()

	// Create and save order
	order, _ := domain.NewOrder(123, s.createValidItems())
	err := s.repo.Save(ctx, order)
	s.Require().NoError(err)
	s.Equal(domain.OrderStatusPending, order.Status)

	// Update to paid
	err = order.MarkAsPaid()
	s.Require().NoError(err)
	err = s.repo.Update(ctx, order)
	s.NoError(err)

	// Verify paid status
	found, err := s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Equal(domain.OrderStatusPaid, found.Status)

	// Update to shipped
	err = found.Ship()
	s.Require().NoError(err)
	err = s.repo.Update(ctx, found)
	s.NoError(err)

	// Verify shipped status
	found, err = s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Equal(domain.OrderStatusShipped, found.Status)

	// Update to completed
	err = found.Complete()
	s.Require().NoError(err)
	err = s.repo.Update(ctx, found)
	s.NoError(err)

	// Verify completed status
	found, err = s.repo.FindByID(ctx, order.ID)
	s.NoError(err)
	s.Equal(domain.OrderStatusCompleted, found.Status)
}
