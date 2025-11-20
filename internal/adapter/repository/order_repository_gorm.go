package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/jmoiron/sqlx"
	"github.com/samber/oops"
)

// orderRepositorySQLX is the sqlx implementation of OrderRepository
type orderRepositorySQLX struct {
	db *sqlx.DB
}

// NewOrderRepository creates a new OrderRepository implementation using sqlx
func NewOrderRepository(db *sqlx.DB) port.OrderRepository {
	return &orderRepositorySQLX{db: db}
}

// FindByID retrieves an order by ID
func (r *orderRepositorySQLX) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	var orderModel OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders WHERE id = ?"
	err := r.db.GetContext(ctx, &orderModel, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeOrderNotFound).
				In("repository").
				Tags("database", "sqlx").
				With("order_id", id).
				Wrap(domain.ErrOrderNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlx").
			With("order_id", id).
			With("query", query).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return toDomainOrder(&orderModel), nil
}

// FindByUserID retrieves all orders for a specific user
func (r *orderRepositorySQLX) FindByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	var orderModels []OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders WHERE user_id = ?"
	err := r.db.SelectContext(ctx, &orderModels, query, userID)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "query").
			With("user_id", userID).
			Wrapf(err, "failed to query orders by user ID")
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = toDomainOrder(&model)
	}

	return orders, nil
}

// Save creates a new order
func (r *orderRepositorySQLX) Save(ctx context.Context, order *domain.Order) error {
	orderModel := toOrderModel(order)
	orderModel.CreatedAt = time.Now().UTC()
	orderModel.UpdatedAt = time.Now().UTC()

	query := `INSERT INTO orders (user_id, total_amount, status, items, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		orderModel.UserID,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.Items,
		orderModel.CreatedAt,
		orderModel.UpdatedAt,
	)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseInsertFailed).
			In("repository").
			Tags("database", "insert").
			With("user_id", orderModel.UserID).
			With("total_amount", orderModel.TotalAmount).
			Hint("Check database connection and constraints").
			Wrapf(err, "failed to insert order")
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database").
			Wrapf(err, "failed to get last insert ID")
	}

	order.ID = id
	order.CreatedAt = orderModel.CreatedAt
	order.UpdatedAt = orderModel.UpdatedAt
	return nil
}

// Update updates an existing order
func (r *orderRepositorySQLX) Update(ctx context.Context, order *domain.Order) error {
	orderModel := toOrderModel(order)
	orderModel.UpdatedAt = time.Now().UTC()

	query := `UPDATE orders
			  SET user_id = ?, total_amount = ?, status = ?, items = ?, updated_at = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		orderModel.UserID,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.Items,
		orderModel.UpdatedAt,
		orderModel.ID,
	)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "update").
			With("order_id", orderModel.ID).
			Wrapf(err, "failed to update order")
	}

	// Check if the order was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database").
			With("order_id", orderModel.ID).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeOrderNotFound).
			In("repository").
			With("order_id", orderModel.ID).
			Hint("Order may have been deleted").
			Wrap(domain.ErrOrderNotFound)
	}

	order.UpdatedAt = orderModel.UpdatedAt
	return nil
}

// UpdateStatus updates the status of an order
func (r *orderRepositorySQLX) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	updatedAt := time.Now().UTC()
	query := "UPDATE orders SET status = ?, updated_at = ? WHERE id = ?"

	result, err := r.db.ExecContext(ctx, query, string(status), updatedAt, orderID)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "status_update").
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
			Tags("database").
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

// List retrieves orders with pagination
func (r *orderRepositorySQLX) List(ctx context.Context, offset, limit int) ([]*domain.Order, error) {
	var orderModels []OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders LIMIT ? OFFSET ?"
	err := r.db.SelectContext(ctx, &orderModels, query, limit, offset)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "query").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to list orders")
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = toDomainOrder(&model)
	}

	return orders, nil
}

// toDomainOrder converts an OrderModel to a domain.Order entity
func toDomainOrder(model *OrderModel) *domain.Order {
	return &domain.Order{
		ID:          model.ID,
		UserID:      model.UserID,
		TotalAmount: model.TotalAmount,
		Status:      domain.OrderStatus(model.Status),
		Items:       []domain.OrderItem(model.Items),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// toOrderModel converts a domain.Order to an OrderModel
func toOrderModel(order *domain.Order) *OrderModel {
	return &OrderModel{
		ID:          order.ID,
		UserID:      order.UserID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		Items:       OrderItemsJSON(order.Items),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}
