package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/jmoiron/sqlx"
	"github.com/samber/oops"
)

// userRepositorySQLX is the sqlx implementation of UserRepository
type userRepositorySQLX struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository implementation using sqlx
func NewUserRepository(db *sqlx.DB) port.UserRepository {
	return &userRepositorySQLX{db: db}
}

// FindByID retrieves a user by ID
func (r *userRepositorySQLX) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var userModel UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?"
	err := r.db.GetContext(ctx, &userModel, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("repository").
				Tags("database", "sqlx").
				With("user_id", id).
				Wrap(domain.ErrUserNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlx").
			With("user_id", id).
			With("query", query).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return toDomainUser(&userModel), nil
}

// FindByEmail retrieves a user by email
func (r *userRepositorySQLX) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var userModel UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?"
	err := r.db.GetContext(ctx, &userModel, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("repository").
				Tags("database", "sqlx").
				With("email", email).
				Wrap(domain.ErrUserNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlx").
			With("email", email).
			With("query", query).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return toDomainUser(&userModel), nil
}

// Save creates a new user
func (r *userRepositorySQLX) Save(ctx context.Context, user *domain.User) error {
	userModel := toUserModel(user)
	userModel.CreatedAt = time.Now().UTC()
	userModel.UpdatedAt = time.Now().UTC()

	query := `INSERT INTO users (email, password_hash, created_at, updated_at)
			  VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		userModel.Email,
		userModel.PasswordHash,
		userModel.CreatedAt,
		userModel.UpdatedAt,
	)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseInsertFailed).
			In("repository").
			Tags("database", "insert").
			With("email", userModel.Email).
			Hint("Check for unique constraint violations (email already exists)").
			Wrapf(err, "failed to insert user")
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database").
			Wrapf(err, "failed to get last insert ID")
	}

	user.ID = id
	user.CreatedAt = userModel.CreatedAt
	user.UpdatedAt = userModel.UpdatedAt
	return nil
}

// Update updates an existing user
func (r *userRepositorySQLX) Update(ctx context.Context, user *domain.User) error {
	userModel := toUserModel(user)
	userModel.UpdatedAt = time.Now().UTC()

	query := `UPDATE users
			  SET email = ?, password_hash = ?, updated_at = ?
			  WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		userModel.Email,
		userModel.PasswordHash,
		userModel.UpdatedAt,
		userModel.ID,
	)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "update").
			With("user_id", userModel.ID).
			Wrapf(err, "failed to update user")
	}

	// Check if the user was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database").
			With("user_id", userModel.ID).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeUserNotFound).
			In("repository").
			With("user_id", userModel.ID).
			Hint("User may have been deleted").
			Wrap(domain.ErrUserNotFound)
	}

	user.UpdatedAt = userModel.UpdatedAt
	return nil
}

// Delete deletes a user by ID
func (r *userRepositorySQLX) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseDeleteFailed).
			In("repository").
			Tags("database", "delete").
			With("user_id", id).
			Wrapf(err, "failed to delete user")
	}

	// Check if the user was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database").
			With("user_id", id).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeUserNotFound).
			In("repository").
			With("user_id", id).
			Hint("User may already be deleted or never existed").
			Wrap(domain.ErrUserNotFound)
	}

	return nil
}

// List retrieves users with pagination
func (r *userRepositorySQLX) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	var userModels []UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users LIMIT ? OFFSET ?"
	err := r.db.SelectContext(ctx, &userModels, query, limit, offset)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "query").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to list users")
	}

	users := make([]*domain.User, len(userModels))
	for i, model := range userModels {
		users[i] = toDomainUser(&model)
	}

	return users, nil
}

// toDomainUser converts a UserModel to a domain.User entity
// Uses ReconstructUser to properly hydrate the entity with the hashed password
func toDomainUser(model *UserModel) *domain.User {
	return domain.ReconstructUser(
		model.ID,
		model.Email,
		model.PasswordHash, // Already hashed
		model.CreatedAt,
		model.UpdatedAt,
	)
}

// toUserModel converts a domain.User to a UserModel
func toUserModel(user *domain.User) *UserModel {
	return &UserModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.Password(), // Use the getter method
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
