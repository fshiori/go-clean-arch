package usecase

import "go-clean-arch/internal/domain"

// UserUsecase defines the interface for user-related business logic
type UserUsecase interface {
	CreateUser(email, password string) (*domain.User, error)
	GetUserByID(id int64) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUserPassword(userID int64, oldPassword, newPassword string) error
	ListUsers(page, pageSize int) ([]*domain.User, error)
	DeleteUser(id int64) error
}
