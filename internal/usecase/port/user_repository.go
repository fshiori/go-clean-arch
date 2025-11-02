package port

import "go-clean-arch/internal/domain"

// UserRepository defines the interface for user data access
// This is defined in the usecase layer, following the dependency inversion principle
type UserRepository interface {
	// FindByID retrieves a user by their ID
	FindByID(id int64) (*domain.User, error)

	// FindByEmail retrieves a user by their email
	FindByEmail(email string) (*domain.User, error)

	// Save creates a new user
	Save(user *domain.User) error

	// Update updates an existing user
	Update(user *domain.User) error

	// Delete deletes a user by ID
	Delete(id int64) error

	// List retrieves users with pagination
	List(offset, limit int) ([]*domain.User, error)
}
