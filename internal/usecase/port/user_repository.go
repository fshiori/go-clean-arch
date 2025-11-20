package port

import (
	"context"
	"go-clean-arch/internal/domain"
)

// UserRepository defines the interface for user data access
// This is defined in the usecase layer, following the dependency inversion principle
type UserRepository interface {
	// FindByID retrieves a user by their ID
	FindByID(ctx context.Context, id int64) (*domain.User, error)

	// FindByEmail retrieves a user by their email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// Save creates a new user
	Save(ctx context.Context, user *domain.User) error

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id int64) error

	// List retrieves users with pagination
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
