package usecase

import (
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
)

// UserInteractor handles user-related business logic
type UserInteractor struct {
	userRepo port.UserRepository
}

// NewUserInteractor creates a new UserInteractor
func NewUserInteractor(userRepo port.UserRepository) *UserInteractor {
	return &UserInteractor{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (i *UserInteractor) CreateUser(email, password string) (*domain.User, error) {
	// Check if email already exists
	existingUser, err := i.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Create new user using domain factory function
	user, err := domain.NewUser(email, password)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := i.userRepo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (i *UserInteractor) GetUserByID(id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := i.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (i *UserInteractor) GetUserByEmail(email string) (*domain.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := i.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserPassword updates a user's password
func (i *UserInteractor) UpdateUserPassword(userID int64, oldPassword, newPassword string) error {
	// Retrieve user
	user, err := i.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Use domain method to change password (includes validation)
	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		return err
	}

	// Update in repository
	if err := i.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

// ListUsers retrieves a list of users with pagination
func (i *UserInteractor) ListUsers(page, pageSize int) ([]*domain.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, err := i.userRepo.List(offset, pageSize)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser deletes a user
func (i *UserInteractor) DeleteUser(id int64) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	return i.userRepo.Delete(id)
}
