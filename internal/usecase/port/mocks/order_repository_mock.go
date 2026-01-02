// Package mocks provides mock implementations of port interfaces for testing.
package mocks

import (
	"context"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/stretchr/testify/mock"
)

// MockOrderRepository is a mock implementation of port.OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

// Ensure MockOrderRepository implements port.OrderRepository
var _ port.OrderRepository = (*MockOrderRepository)(nil)

// FindByID mocks the FindByID method
func (m *MockOrderRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).(*domain.Order), args.Error(1)
}

// FindByUserID mocks the FindByUserID method
func (m *MockOrderRepository) FindByUserID(ctx context.Context, userID int64) ([]*domain.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).([]*domain.Order), args.Error(1)
}

// Save mocks the Save method
func (m *MockOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// UpdateStatus mocks the UpdateStatus method
func (m *MockOrderRepository) UpdateStatus(ctx context.Context, orderID int64, status domain.OrderStatus) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

// List mocks the List method
func (m *MockOrderRepository) List(ctx context.Context, offset, limit int) ([]*domain.Order, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).([]*domain.Order), args.Error(1)
}
