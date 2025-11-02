package repository

import (
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"gorm.io/gorm"
)

// orderRepositoryGORM is the GORM implementation of OrderRepository
type orderRepositoryGORM struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository implementation
func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &orderRepositoryGORM{db: db}
}

// FindByID retrieves an order by ID
func (r *orderRepositoryGORM) FindByID(id int64) (*domain.Order, error) {
	var orderModel OrderModel

	if err := r.db.First(&orderModel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	return toDomainOrder(&orderModel), nil
}

// FindByUserID retrieves all orders for a specific user
func (r *orderRepositoryGORM) FindByUserID(userID int64) ([]*domain.Order, error) {
	var orderModels []OrderModel

	if err := r.db.Where("user_id = ?", userID).Find(&orderModels).Error; err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, len(orderModels))
	for i, model := range orderModels {
		orders[i] = toDomainOrder(&model)
	}

	return orders, nil
}

// Save creates a new order
func (r *orderRepositoryGORM) Save(order *domain.Order) error {
	orderModel := toOrderModel(order)

	if err := r.db.Create(orderModel).Error; err != nil {
		return err
	}

	// Update the domain entity with the generated ID
	order.ID = orderModel.ID

	return nil
}

// Update updates an existing order
func (r *orderRepositoryGORM) Update(order *domain.Order) error {
	orderModel := toOrderModel(order)

	if err := r.db.Save(orderModel).Error; err != nil {
		return err
	}

	return nil
}

// UpdateStatus updates the status of an order
func (r *orderRepositoryGORM) UpdateStatus(orderID int64, status domain.OrderStatus) error {
	if err := r.db.Model(&OrderModel{}).Where("id = ?", orderID).Update("status", string(status)).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves orders with pagination
func (r *orderRepositoryGORM) List(offset, limit int) ([]*domain.Order, error) {
	var orderModels []OrderModel

	if err := r.db.Offset(offset).Limit(limit).Find(&orderModels).Error; err != nil {
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
