package repository

import (
	"database/sql"
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
	"time"

	"github.com/jmoiron/sqlx"
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
func (r *orderRepositorySQLX) FindByID(id int64) (*domain.Order, error) {
	var orderModel OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders WHERE id = ?"
	err := r.db.Get(&orderModel, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	return toDomainOrder(&orderModel), nil
}

// FindByUserID retrieves all orders for a specific user
func (r *orderRepositorySQLX) FindByUserID(userID int64) ([]*domain.Order, error) {
	var orderModels []OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders WHERE user_id = ?"
	err := r.db.Select(&orderModels, query, userID)
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = toDomainOrder(&model)
	}

	return orders, nil
}

// Save creates a new order
func (r *orderRepositorySQLX) Save(order *domain.Order) error {
	orderModel := toOrderModel(order)
	orderModel.CreatedAt = time.Now().UTC()
	orderModel.UpdatedAt = time.Now().UTC()

	query := `INSERT INTO orders (user_id, total_amount, status, items, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		orderModel.UserID,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.Items,
		orderModel.CreatedAt,
		orderModel.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	order.ID = id
	return nil
}

// Update updates an existing order
func (r *orderRepositorySQLX) Update(order *domain.Order) error {
	orderModel := toOrderModel(order)
	orderModel.UpdatedAt = time.Now().UTC()

	query := `UPDATE orders
			  SET user_id = ?, total_amount = ?, status = ?, items = ?, updated_at = ?
			  WHERE id = ?`

	_, err := r.db.Exec(query,
		orderModel.UserID,
		orderModel.TotalAmount,
		orderModel.Status,
		orderModel.Items,
		orderModel.UpdatedAt,
		orderModel.ID,
	)

	return err
}

// UpdateStatus updates the status of an order
func (r *orderRepositorySQLX) UpdateStatus(orderID int64, status domain.OrderStatus) error {
	query := "UPDATE orders SET status = ?, updated_at = ? WHERE id = ?"
	_, err := r.db.Exec(query, string(status), time.Now().UTC(), orderID)
	return err
}

// List retrieves orders with pagination
func (r *orderRepositorySQLX) List(offset, limit int) ([]*domain.Order, error) {
	var orderModels []OrderModel

	query := "SELECT id, user_id, total_amount, status, items, created_at, updated_at FROM orders LIMIT ? OFFSET ?"
	err := r.db.Select(&orderModels, query, limit, offset)
	if err != nil {
		return nil, err
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
