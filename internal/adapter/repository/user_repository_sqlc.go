package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-clean-arch/internal/adapter/repository/sqlcgen"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/samber/oops"
)

// MySQL uses ? placeholders
var mysql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Question)

// userRepositorySQLC combines sqlc (for static queries) and sqlx+Squirrel (for dynamic queries)
// This follows the "sqlc primary, sqlx+Squirrel auxiliary" approach defined in CODING_STANDARDS.md
type userRepositorySQLC struct {
	db      *sqlx.DB
	queries *sqlcgen.Queries
}

// NewUserRepository creates a new UserRepository implementation using sqlc + sqlx
func NewUserRepository(db *sqlx.DB) port.UserRepository {
	return &userRepositorySQLC{
		db:      db,
		queries: sqlcgen.New(db.DB),
	}
}

// FindByID retrieves a user by ID using sqlc (static query)
func (r *userRepositorySQLC) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("repository").
				Tags("database", "sqlc").
				With("user_id", id).
				Wrap(domain.ErrUserNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("user_id", id).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return sqlcUserToDomain(user), nil
}

// FindByEmail retrieves a user by email using sqlc (static query)
func (r *userRepositorySQLC) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, oops.
				Code(domain.ErrCodeUserNotFound).
				In("repository").
				Tags("database", "sqlc").
				With("email", email).
				Wrap(domain.ErrUserNotFound)
		}

		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("email", email).
			Hint("Check database connection and schema").
			Wrapf(err, "database query failed")
	}

	return sqlcUserToDomain(user), nil
}

// Save creates a new user using sqlc (static query)
func (r *userRepositorySQLC) Save(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC()

	result, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.Password(),
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseInsertFailed).
			In("repository").
			Tags("database", "sqlc", "insert").
			With("email", user.Email).
			Hint("Check for unique constraint violations (email already exists)").
			Wrapf(err, "failed to insert user")
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			Wrapf(err, "failed to get last insert ID")
	}

	user.ID = id
	user.CreatedAt = now
	user.UpdatedAt = now
	return nil
}

// Update updates an existing user using sqlc (static query)
func (r *userRepositorySQLC) Update(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC()

	result, err := r.queries.UpdateUser(ctx, sqlcgen.UpdateUserParams{
		Email:        user.Email,
		PasswordHash: user.Password(),
		UpdatedAt:    now,
		ID:           user.ID,
	})
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseUpdateFailed).
			In("repository").
			Tags("database", "sqlc", "update").
			With("user_id", user.ID).
			Wrapf(err, "failed to update user")
	}

	// Check if the user was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
			With("user_id", user.ID).
			Wrapf(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return oops.
			Code(domain.ErrCodeUserNotFound).
			In("repository").
			With("user_id", user.ID).
			Hint("User may have been deleted").
			Wrap(domain.ErrUserNotFound)
	}

	user.UpdatedAt = now
	return nil
}

// Delete deletes a user by ID using sqlc (static query)
func (r *userRepositorySQLC) Delete(ctx context.Context, id int64) error {
	result, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseDeleteFailed).
			In("repository").
			Tags("database", "sqlc", "delete").
			With("user_id", id).
			Wrapf(err, "failed to delete user")
	}

	// Check if the user was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "sqlc").
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

// List retrieves users with pagination using sqlx+Squirrel (dynamic query)
// This demonstrates the "auxiliary" use case: dynamic queries with flexible parameters
func (r *userRepositorySQLC) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	// Build dynamic query with Squirrel
	query, args, err := mysql.
		Select("id", "email", "password_hash", "created_at", "updated_at").
		From("users").
		OrderBy("created_at DESC"). // Could be dynamic based on sort parameters
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		ToSql()

	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseError).
			In("repository").
			Tags("database", "squirrel", "query-builder").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to build query")
	}

	// Execute with sqlx
	var sqlcUsers []sqlcgen.User
	err = r.db.SelectContext(ctx, &sqlcUsers, query, args...)
	if err != nil {
		return nil, oops.
			Code(domain.ErrCodeDatabaseQueryFailed).
			In("repository").
			Tags("database", "sqlx", "squirrel").
			With("offset", offset).
			With("limit", limit).
			Wrapf(err, "failed to list users")
	}

	// Convert to domain entities
	users := make([]*domain.User, len(sqlcUsers))
	for i, sqlcUser := range sqlcUsers {
		users[i] = sqlcUserToDomain(&sqlcUser)
	}

	return users, nil
}

// sqlcUserToDomain converts a sqlc-generated User to a domain.User entity
// Uses ReconstructUser to properly hydrate the entity with the hashed password
func sqlcUserToDomain(sqlcUser *sqlcgen.User) *domain.User {
	return domain.ReconstructUser(
		sqlcUser.ID,
		sqlcUser.Email,
		sqlcUser.PasswordHash, // Already hashed
		sqlcUser.CreatedAt,
		sqlcUser.UpdatedAt,
	)
}
