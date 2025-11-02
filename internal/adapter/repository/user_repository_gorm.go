package repository

import (
	"errors"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"gorm.io/gorm"
)

// userRepositoryGORM is the GORM implementation of UserRepository
type userRepositoryGORM struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository implementation
func NewUserRepository(db *gorm.DB) port.UserRepository {
	return &userRepositoryGORM{db: db}
}

// FindByID retrieves a user by ID
func (r *userRepositoryGORM) FindByID(id int64) (*domain.User, error) {
	var userModel UserModel

	if err := r.db.First(&userModel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Convert DB Model to Domain Entity
	return toDomainUser(&userModel), nil
}

// FindByEmail retrieves a user by email
func (r *userRepositoryGORM) FindByEmail(email string) (*domain.User, error) {
	var userModel UserModel

	if err := r.db.Where("email = ?", email).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return toDomainUser(&userModel), nil
}

// Save creates a new user
func (r *userRepositoryGORM) Save(user *domain.User) error {
	userModel := toUserModel(user)

	if err := r.db.Create(userModel).Error; err != nil {
		return err
	}

	// Update the domain entity with the generated ID
	user.ID = userModel.ID

	return nil
}

// Update updates an existing user
func (r *userRepositoryGORM) Update(user *domain.User) error {
	userModel := toUserModel(user)

	if err := r.db.Save(userModel).Error; err != nil {
		return err
	}

	return nil
}

// Delete deletes a user by ID
func (r *userRepositoryGORM) Delete(id int64) error {
	if err := r.db.Delete(&UserModel{}, id).Error; err != nil {
		return err
	}

	return nil
}

// List retrieves users with pagination
func (r *userRepositoryGORM) List(offset, limit int) ([]*domain.User, error) {
	var userModels []UserModel

	if err := r.db.Offset(offset).Limit(limit).Find(&userModels).Error; err != nil {
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
