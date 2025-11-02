package domain

import (
	"errors"
	"time"
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
	// Validation
	if email == "" {
		return nil, errors.New("email is required")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password (simplified for example)
	hashedPassword := hashPassword(password)

	return &User{
		Email:     email,
		password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// Password returns the hashed password
func (u *User) Password() string {
	return u.password
}

// ChangePassword changes the user's password with validation
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	if !u.IsPasswordCorrect(oldPassword) {
		return errors.New("incorrect old password")
	}
	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	u.password = hashPassword(newPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// IsPasswordCorrect checks if the provided password matches
func (u *User) IsPasswordCorrect(password string) bool {
	return comparePassword(u.password, password)
}

// hashPassword is a simplified password hashing function
// In production, use bcrypt or similar
func hashPassword(password string) string {
	// TODO: implement proper password hashing (bcrypt, argon2, etc.)
	return "hashed_" + password
}

// comparePassword is a simplified password comparison function
func comparePassword(hashedPassword, password string) bool {
	// TODO: implement proper password comparison
	return hashedPassword == "hashed_"+password
}
