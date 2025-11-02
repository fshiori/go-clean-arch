package repository

import "time"

// UserModel is the database model for users
// This struct is specific to the database layer and contains GORM tags
// It is separate from the domain.User entity
type UserModel struct {
	ID           int64     `gorm:"column:id;primary_key;autoIncrement"`
	Email        string    `gorm:"column:email;unique;not null"`
	PasswordHash string    `gorm:"column:password_hash;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

// TableName specifies the table name for GORM
func (UserModel) TableName() string {
	return "users"
}
