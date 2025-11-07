package repository

import "time"

// UserModel is the database model for users
// This struct is specific to the database layer and contains db tags for sqlx
// It is separate from the domain.User entity
type UserModel struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
