package domain

import (
	"time"

	"github.com/samber/oops"
	"golang.org/x/crypto/bcrypt"
)

// User represents a core business entity
// It contains no tags (no json, no gorm) and is completely independent
type User struct {
	ID        int64
	Email     string
	password  string // private field, encapsulated
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser is a factory function to create a new User
// This ensures that every User created is in a valid state
func NewUser(email, password string) (*User, error) {
	// Validation - Email
	if email == "" {
		return nil, oops.
			Code(ErrCodeInvalidEmail).
			In("domain").
			With("email", email).
			Hint("Email cannot be empty").
			Wrap(ErrInvalidEmail)
	}

	// Validation - Password length
	if len(password) < 8 {
		return nil, oops.
			Code(ErrCodePasswordTooShort).
			In("domain").
			With("password_length", len(password)).
			Hint("Password must be at least 8 characters").
			Wrap(ErrPasswordTooShort)
	}

	// Hash password with bcrypt
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		Email:     email,
		password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ReconstructUser creates a User from persisted data (for repository use only)
// This method is used to restore User entities from the database without re-hashing the password
// Parameters:
//   - passwordHash: should already be a hashed value (bcrypt hash)
func ReconstructUser(id int64, email, passwordHash string, createdAt, updatedAt time.Time) *User {
	return &User{
		ID:        id,
		Email:     email,
		password:  passwordHash, // Already hashed
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

// Password returns the hashed password
func (u *User) Password() string {
	return u.password
}

// ChangePassword changes the user's password with validation
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	// Verify old password
	if !u.IsPasswordCorrect(oldPassword) {
		return oops.
			Code(ErrCodeIncorrectPassword).
			In("domain").
			With("user_id", u.ID).
			Hint("Old password doesn't match").
			Wrap(ErrIncorrectPassword)
	}

	// Validate new password length
	if len(newPassword) < 8 {
		return oops.
			Code(ErrCodePasswordTooShort).
			In("domain").
			With("user_id", u.ID).
			With("password_length", len(newPassword)).
			Hint("New password must be at least 8 characters").
			Wrap(ErrPasswordTooShort)
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return oops.
			In("domain").
			With("user_id", u.ID).
			Wrapf(err, "failed to hash new password")
	}

	u.password = hashedPassword
	u.UpdatedAt = time.Now()
	return nil
}

// IsPasswordCorrect checks if the provided password matches
func (u *User) IsPasswordCorrect(password string) bool {
	return comparePassword(u.password, password)
}

// hashPassword uses bcrypt to hash the password
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", oops.
			Code(ErrCodePasswordHashFailed).
			In("domain").
			Hint("Failed to hash password with bcrypt").
			Wrapf(err, "bcrypt hash failed")
	}
	return string(hash), nil
}

// comparePassword compares a hashed password with a plain text password
func comparePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
