package repository

import (
	"database/sql/driver"
	"encoding/json"
	"go-clean-arch/internal/domain"
	"time"
)

// OrderModel is the database model for orders
type OrderModel struct {
	ID          int64              `gorm:"column:id;primary_key;autoIncrement"`
	UserID      int64              `gorm:"column:user_id;not null;index"`
	TotalAmount float64            `gorm:"column:total_amount;not null"`
	Status      string             `gorm:"column:status;not null"`
	Items       OrderItemsJSON     `gorm:"column:items;type:json"`
	CreatedAt   time.Time          `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time          `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for GORM
func (OrderModel) TableName() string {
	return "orders"
}

// OrderItemsJSON is a custom type for storing order items as JSON in the database
type OrderItemsJSON []domain.OrderItem

// Scan implements the sql.Scanner interface for reading from the database
func (o *OrderItemsJSON) Scan(value interface{}) error {
	if value == nil {
		*o = OrderItemsJSON{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	var items []domain.OrderItem
	if err := json.Unmarshal(bytes, &items); err != nil {
		return err
	}

	*o = items
	return nil
}

// Value implements the driver.Valuer interface for writing to the database
func (o OrderItemsJSON) Value() (driver.Value, error) {
	if len(o) == 0 {
		return "[]", nil
	}

	bytes, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	return string(bytes), nil
}
