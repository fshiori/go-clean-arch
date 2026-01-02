// Package mocks provides mock implementations of port interfaces for testing.
package mocks

import (
	"context"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of port.UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Ensure MockUserRepository implements port.UserRepository
var _ port.UserRepository = (*MockUserRepository)(nil)

// FindByID mocks the FindByID method
func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).(*domain.User), args.Error(1)
}

// FindByEmail mocks the FindByEmail method
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).(*domain.User), args.Error(1)
}

// Save mocks the Save method
func (m *MockUserRepository) Save(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks the List method
func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	//nolint:errcheck // testify mock pattern
	return args.Get(0).([]*domain.User), args.Error(1)
}
