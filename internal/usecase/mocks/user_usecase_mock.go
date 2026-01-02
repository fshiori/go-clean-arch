package mocks

import (
	"context"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase"

	"github.com/stretchr/testify/mock"
)

// MockUserUsecase is a mock implementation of usecase.UserUsecase
type MockUserUsecase struct {
	mock.Mock
}

// Ensure MockUserUsecase implements usecase.UserUsecase
var _ usecase.UserUsecase = (*MockUserUsecase)(nil)

// CreateUser mocks the CreateUser method
func (m *MockUserUsecase) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetUserByID mocks the GetUserByID method
func (m *MockUserUsecase) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// GetUserByEmail mocks the GetUserByEmail method
func (m *MockUserUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// UpdateUserPassword mocks the UpdateUserPassword method
func (m *MockUserUsecase) UpdateUserPassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

// DeleteUser mocks the DeleteUser method
func (m *MockUserUsecase) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ListUsers mocks the ListUsers method
func (m *MockUserUsecase) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}
