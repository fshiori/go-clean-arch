package usecase

import (
	"context"
	"errors"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/samber/oops"
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
func (i *UserInteractor) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
	// Check if email already exists
	existingUser, err := i.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, oops.
			Code(domain.ErrCodeEmailAlreadyExists).
			In("user_usecase").
			Tags("registration", "validation").
			With("email", email).
			Hint("Try using a different email address").
			Wrap(domain.ErrEmailAlreadyExists)
	}

	// Create new user using domain factory function
	user, err := domain.NewUser(email, password)
	if err != nil {
		return nil, oops.
			In("user_usecase").
			Tags("registration", "validation").
			With("email", email).
			Wrapf(err, "failed to create user entity")
	}

	// Save to repository
	if err := i.userRepo.Save(ctx, user); err != nil {
		return nil, oops.
			Code(domain.ErrCodeUserSaveFailed).
			In("user_usecase").
			Tags("registration", "database").
			With("email", email).
			Hint("Check database connection and constraints").
			Wrapf(err, "failed to save user to repository")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (i *UserInteractor) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if id <= 0 {
		return nil, oops.
			Code(domain.ErrCodeInvalidUserID).
			In("user_usecase").
			With("user_id", id).
			Hint("User ID must be a positive integer").
			Wrap(domain.ErrInvalidUserID)
	}

	user, err := i.userRepo.FindByID(ctx, id)
	if err != nil {
		// If it's a "not found" error, use domain error code
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("user_usecase").
				With("user_id", id).
				Wrap(err)
		}

		// Other database errors
		return nil, oops.
			Code(domain.ErrCodeUserFetchFailed).
			In("user_usecase").
			With("user_id", id).
			Hint("Check database connection").
			Wrapf(err, "failed to fetch user from repository")
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (i *UserInteractor) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, oops.
			Code(domain.ErrCodeInvalidEmail).
			In("user_usecase").
			With("email", email).
			Wrap(domain.ErrInvalidEmail)
	}

	user, err := i.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// If it's a "not found" error, use domain error code
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("user_usecase").
				With("email", email).
				Wrap(err)
		}

		// Other database errors
		return nil, oops.
			Code(domain.ErrCodeUserFetchFailed).
			In("user_usecase").
			With("email", email).
			Hint("Check database connection").
			Wrapf(err, "failed to fetch user from repository")
	}

	return user, nil
}

// UpdateUserPassword updates a user's password
func (i *UserInteractor) UpdateUserPassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// Retrieve user
	user, err := i.GetUserByID(ctx, userID)
	if err != nil {
		return oops.
			In("user_usecase").
			Tags("password_update").
			Wrapf(err, "failed to retrieve user for password update")
	}

	// Use domain method to change password (includes validation)
	if err := user.ChangePassword(oldPassword, newPassword); err != nil {
		return oops.
			In("user_usecase").
			Tags("password_update", "validation").
			With("user_id", userID).
			Wrapf(err, "failed to change password")
	}

	// Update in repository
	if err := i.userRepo.Update(ctx, user); err != nil {
		return oops.
			Code(domain.ErrCodeUserUpdateFailed).
			In("user_usecase").
			Tags("password_update", "database").
			With("user_id", userID).
			Hint("Check database connection").
			Wrapf(err, "failed to update user in repository")
	}

	return nil
}

// ListUsers retrieves a list of users with pagination
func (i *UserInteractor) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, err := i.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeUserListFailed).
			In("user_usecase").
			Tags("listing").
			With("page", page).
			With("page_size", pageSize).
			Hint("Check database connection").
			Wrapf(err, "failed to list users")
	}

	return users, nil
}

// DeleteUser deletes a user
func (i *UserInteractor) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return oops.
			Code(domain.ErrCodeInvalidUserID).
			In("user_usecase").
			With("user_id", id).
			Wrap(domain.ErrInvalidUserID)
	}

	if err := i.userRepo.Delete(ctx, id); err != nil {
		return oops.
			Code(domain.ErrCodeUserDeleteFailed).
			In("user_usecase").
			Tags("deletion").
			With("user_id", id).
			Hint("Check if user exists and database connection").
			Wrapf(err, "failed to delete user")
	}

	return nil
}
