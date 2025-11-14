package repository

import (
	"context"
	"testing"
	"time"

	"go-clean-arch/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/suite"
)

type UserRepositorySQLXTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	repo *userRepositorySQLX
}

func TestUserRepositorySQLXTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySQLXTestSuite))
}

func (s *UserRepositorySQLXTestSuite) SetupSuite() {
	// Create in-memory SQLite database
	db, err := sqlx.Connect("sqlite3", ":memory:")
	s.Require().NoError(err)

	// Create users table
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	_, err = db.Exec(schema)
	s.Require().NoError(err)

	s.db = db
	s.repo = &userRepositorySQLX{db: db}
}

func (s *UserRepositorySQLXTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *UserRepositorySQLXTestSuite) TearDownTest() {
	// Clean up users table after each test
	_, err := s.db.Exec("DELETE FROM users")
	s.Require().NoError(err)
}

// Test Save
func (s *UserRepositorySQLXTestSuite) TestSave_Success() {
	ctx := context.Background()
	user, err := domain.NewUser("test@example.com", "password123")
	s.Require().NoError(err)

	// Save user
	err = s.repo.Save(ctx, user)
	s.NoError(err)
	s.NotZero(user.ID)
	s.NotZero(user.CreatedAt)
	s.NotZero(user.UpdatedAt)
}

func (s *UserRepositorySQLXTestSuite) TestSave_DuplicateEmail() {
	ctx := context.Background()
	user1, _ := domain.NewUser("duplicate@example.com", "password123")
	user2, _ := domain.NewUser("duplicate@example.com", "password456")

	// Save first user
	err := s.repo.Save(ctx, user1)
	s.NoError(err)

	// Try to save second user with same email
	err = s.repo.Save(ctx, user2)
	s.Error(err)
	s.Contains(err.Error(), "failed to insert user")
}

// Test FindByID
func (s *UserRepositorySQLXTestSuite) TestFindByID_Success() {
	ctx := context.Background()
	user, _ := domain.NewUser("find@example.com", "password123")

	// Save user first
	err := s.repo.Save(ctx, user)
	s.Require().NoError(err)

	// Find by ID
	found, err := s.repo.FindByID(ctx, user.ID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID, found.ID)
	s.Equal(user.Email, found.Email)
}

func (s *UserRepositorySQLXTestSuite) TestFindByID_NotFound() {
	ctx := context.Background()

	// Try to find non-existent user
	found, err := s.repo.FindByID(ctx, 99999)
	s.Error(err)
	s.Nil(found)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

// Test FindByEmail
func (s *UserRepositorySQLXTestSuite) TestFindByEmail_Success() {
	ctx := context.Background()
	user, _ := domain.NewUser("email@example.com", "password123")

	// Save user first
	err := s.repo.Save(ctx, user)
	s.Require().NoError(err)

	// Find by email
	found, err := s.repo.FindByEmail(ctx, "email@example.com")
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID, found.ID)
	s.Equal(user.Email, found.Email)
}

func (s *UserRepositorySQLXTestSuite) TestFindByEmail_NotFound() {
	ctx := context.Background()

	// Try to find non-existent user
	found, err := s.repo.FindByEmail(ctx, "nonexistent@example.com")
	s.Error(err)
	s.Nil(found)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

// Test Update
func (s *UserRepositorySQLXTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	user, _ := domain.NewUser("update@example.com", "password123")

	// Save user first
	err := s.repo.Save(ctx, user)
	s.Require().NoError(err)

	// Update password
	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	err = user.UpdatePassword("password123", "newpassword456")
	s.Require().NoError(err)

	err = s.repo.Update(ctx, user)
	s.NoError(err)
	s.True(user.UpdatedAt.After(originalUpdatedAt))

	// Verify update
	found, err := s.repo.FindByID(ctx, user.ID)
	s.NoError(err)
	s.True(found.VerifyPassword("newpassword456"))
}

func (s *UserRepositorySQLXTestSuite) TestUpdate_NotFound() {
	ctx := context.Background()
	user := domain.ReconstructUser(99999, "nonexistent@example.com", "hash", time.Now(), time.Now())

	// Try to update non-existent user
	err := s.repo.Update(ctx, user)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

// Test Delete
func (s *UserRepositorySQLXTestSuite) TestDelete_Success() {
	ctx := context.Background()
	user, _ := domain.NewUser("delete@example.com", "password123")

	// Save user first
	err := s.repo.Save(ctx, user)
	s.Require().NoError(err)

	// Delete user
	err = s.repo.Delete(ctx, user.ID)
	s.NoError(err)

	// Verify deletion
	_, err = s.repo.FindByID(ctx, user.ID)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserRepositorySQLXTestSuite) TestDelete_NotFound() {
	ctx := context.Background()

	// Try to delete non-existent user
	err := s.repo.Delete(ctx, 99999)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

// Test List
func (s *UserRepositorySQLXTestSuite) TestList_Success() {
	ctx := context.Background()

	// Create multiple users
	for i := 1; i <= 5; i++ {
		user, _ := domain.NewUser("user"+string(rune('0'+i))+"@example.com", "password123")
		err := s.repo.Save(ctx, user)
		s.Require().NoError(err)
	}

	// List users with pagination
	users, err := s.repo.List(ctx, 0, 3)
	s.NoError(err)
	s.Len(users, 3)

	// List next page
	users, err = s.repo.List(ctx, 3, 3)
	s.NoError(err)
	s.Len(users, 2)
}

func (s *UserRepositorySQLXTestSuite) TestList_Empty() {
	ctx := context.Background()

	// List from empty table
	users, err := s.repo.List(ctx, 0, 10)
	s.NoError(err)
	s.Empty(users)
}
