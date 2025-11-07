package repository

import (
	"database/sql"
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
	"time"

	"github.com/jmoiron/sqlx"
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
func (r *userRepositorySQLX) FindByID(id int64) (*domain.User, error) {
	var userModel UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?"
	err := r.db.Get(&userModel, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return toDomainUser(&userModel), nil
}

// FindByEmail retrieves a user by email
func (r *userRepositorySQLX) FindByEmail(email string) (*domain.User, error) {
	var userModel UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?"
	err := r.db.Get(&userModel, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return toDomainUser(&userModel), nil
}

// Save creates a new user
func (r *userRepositorySQLX) Save(user *domain.User) error {
	userModel := toUserModel(user)
	userModel.CreatedAt = time.Now().UTC()
	userModel.UpdatedAt = time.Now().UTC()

	query := `INSERT INTO users (email, password_hash, created_at, updated_at)
			  VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		userModel.Email,
		userModel.PasswordHash,
		userModel.CreatedAt,
		userModel.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Get the generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

// Update updates an existing user
func (r *userRepositorySQLX) Update(user *domain.User) error {
	userModel := toUserModel(user)
	userModel.UpdatedAt = time.Now().UTC()

	query := `UPDATE users
			  SET email = ?, password_hash = ?, updated_at = ?
			  WHERE id = ?`

	_, err := r.db.Exec(query,
		userModel.Email,
		userModel.PasswordHash,
		userModel.UpdatedAt,
		userModel.ID,
	)

	return err
}

// Delete deletes a user by ID
func (r *userRepositorySQLX) Delete(id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves users with pagination
func (r *userRepositorySQLX) List(offset, limit int) ([]*domain.User, error) {
	var userModels []UserModel

	query := "SELECT id, email, password_hash, created_at, updated_at FROM users LIMIT ? OFFSET ?"
	err := r.db.Select(&userModels, query, limit, offset)
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(userModels))
	for i, model := range userModels {
		users[i] = toDomainUser(&model)
	}

	return users, nil
}

// toDomainUser converts a UserModel to a domain.User entity
// This is a critical transformation that isolates the domain from database details
func toDomainUser(model *UserModel) *domain.User {
	// Note: We need to use a method that can set the private password field
	// In a real implementation, you might need to adjust the domain.User struct
	// or provide a special constructor for repository use
	user := &domain.User{
		ID:        model.ID,
		Email:     model.Email,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	// The password field is private, so we need a way to set it
	// This is a simplified approach; in production, you might use reflection
	// or provide a special method in the domain package
	return user
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
