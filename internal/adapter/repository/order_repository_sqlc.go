package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-clean-arch/internal/adapter/repository/sqlcgen"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/jmoiron/sqlx"
	"github.com/samber/oops"
)

// orderRepositorySQLC combines sqlc (for static queries) and sqlx+Squirrel (for dynamic queries)
// This follows the "sqlc primary, sqlx+Squirrel auxiliary" approach defined in CODING_STANDARDS.md
type orderRepositorySQLC struct {
	db      *sqlx.DB
	queries *sqlcgen.Queries
}

// NewOrderRepository creates a new OrderRepository implementation using sqlc + sqlx
func NewOrderRepository(db *sqlx.DB) port.OrderRepository {
	return &orderRepositorySQLC{
		db:      db,
		queries: sqlcgen.New(db.DB),
	}
}

// FindByID retrieves an order by ID using sqlc (static query)
func (r *orderRepositorySQLC) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	order, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeOrderNotFound).
				In("repository").
				Tags("database", "sqlc").
				With("order_id", id).
				Wrap(domain.ErrOrderNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("order_id", id).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return sqlcOrderToDomain(order), nil
}

// FindByUserID retrieves all orders for a specific user using sqlc (static query)
func (r *orderRepositorySQLC) FindByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	orders, err := r.queries.GetOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "sqlc").
			With("user_id", userID).
			Wrapf(err, "failed to query orders by user ID")
	}

	result := make([]*domain.Order, len(orders))
	for i := range orders {
		result[i] = sqlcOrderToDomain(orders[i])
	}

	return result, nil
}

// Save creates a new order using sqlc (static query)
func (r *orderRepositorySQLC) Save(ctx context.Context, order *domain.Order) error {
	now := time.Now().UTC()

	// Convert domain items to JSON
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseInsertFailed).
			In("repository").
			Tags("json", "marshal").
			Wrapf(err, "failed to marshal order items")
	}

	result, err := r.queries.CreateOrder(ctx, sqlcgen.CreateOrderParams{
		UserID:      order.UserID,
		TotalAmount: fmt.Sprintf("%.2f", order.TotalAmount),
		Status:      string(order.Status),
		Items:       itemsJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseInsertFailed).
			In("repository").
			Tags("database", "sqlc", "insert").
			With("user_id", order.UserID).
			With("total_amount", order.TotalAmount).
			Hint("Check database connection and constraints").
			Wrapf(err, "failed to insert order")
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			Wrapf(err, "failed to get last insert ID")
	}

	order.ID = id
	order.CreatedAt = now
	order.UpdatedAt = now
	return nil
}

// Update updates an existing order using sqlc (static query)
func (r *orderRepositorySQLC) Update(ctx context.Context, order *domain.Order) error {
	now := time.Now().UTC()

	// Convert domain items to JSON
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("json", "marshal").
			Wrapf(err, "failed to marshal order items")
	}

	result, err := r.queries.UpdateOrder(ctx, sqlcgen.UpdateOrderParams{
		UserID:      order.UserID,
		TotalAmount: fmt.Sprintf("%.2f", order.TotalAmount),
		Status:      string(order.Status),
		Items:       itemsJSON,
		UpdatedAt:   now,
		ID:          order.ID,
	})
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "sqlc", "update").
			With("order_id", order.ID).
			Wrapf(err, "failed to update order")
	}

	// Check if the order was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("order_id", order.ID).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeOrderNotFound).
			In("repository").
			With("order_id", order.ID).
			Hint("Order may have been deleted").
			Wrap(domain.ErrOrderNotFound)
	}

	order.UpdatedAt = now
	return nil
}

// UpdateStatus updates the status of an order using sqlc (static query)
func (r *orderRepositorySQLC) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	now := time.Now().UTC()

	result, err := r.queries.UpdateOrderStatus(ctx, sqlcgen.UpdateOrderStatusParams{
		Status:    string(status),
		UpdatedAt: now,
		ID:        orderID,
	})
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "sqlc", "status_update").
			With("order_id", orderID).
			With("status", status).
			Wrapf(err, "failed to update order status")
	}

	// Check if the order was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("order_id", orderID).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeOrderNotFound).
			In("repository").
			With("order_id", orderID).
			Hint("Order may have been deleted").
			Wrap(domain.ErrOrderNotFound)
	}

	return nil
}

// List retrieves orders with pagination using sqlx+Squirrel (dynamic query)
// This demonstrates the "auxiliary" use case: dynamic queries with flexible parameters
func (r *orderRepositorySQLC) List(ctx context.Context, offset, limit int) ([]*domain.Order, error) {
	// Build dynamic query with Squirrel
	query, args, err := mysql.
		Select("id", "user_id", "total_amount", "status", "items", "created_at", "updated_at").
		From("orders").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "squirrel", "query-builder").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to build query")
	}

	// Execute with sqlx
	var sqlcOrders []sqlcgen.Order
	err = r.db.SelectContext(ctx, &sqlcOrders, query, args...)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "sqlx", "squirrel").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to list orders")
	}

	// Convert to domain entities
	orders := make([]*domain.Order, len(sqlcOrders))
	for i, sqlcOrder := range sqlcOrders {
		orders[i] = sqlcOrderToDomain(&sqlcOrder)
	}

	return orders, nil
}

// sqlcOrderToDomain converts a sqlc-generated Order to a domain.Order entity
func sqlcOrderToDomain(sqlcOrder *sqlcgen.Order) *domain.Order {
	// Unmarshal JSON items
	var items []domain.OrderItem
	if len(sqlcOrder.Items) > 0 {
		_ = json.Unmarshal(sqlcOrder.Items, &items)
	}

	// Parse TotalAmount from string to float64
	var totalAmount float64
	fmt.Sscanf(sqlcOrder.TotalAmount, "%f", &totalAmount)

	return &domain.Order{
		ID:          sqlcOrder.ID,
		UserID:      sqlcOrder.UserID,
		TotalAmount: totalAmount,
		Status:      domain.OrderStatus(sqlcOrder.Status),
		Items:       items,
		CreatedAt:   sqlcOrder.CreatedAt,
		UpdatedAt:   sqlcOrder.UpdatedAt,
	}
}
