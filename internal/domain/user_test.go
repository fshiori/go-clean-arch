package domain

import (
	"testing"
	"time"

	"github.com/samber/oops"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

// Test NewUser function
func (s *UserTestSuite) TestNewUser_Success() {
	email := "test@example.com"
	password := "securePassword123"

	user, err := NewUser(email, password)

	s.NoError(err)
	s.NotNil(user)
	s.Equal(email, user.Email)
	s.NotEmpty(user.password)
	s.NotEqual(password, user.password) // Should be hashed
	s.NotZero(user.CreatedAt)
	s.NotZero(user.UpdatedAt)
	s.Equal(int64(0), user.ID) // Not yet saved
}

func (s *UserTestSuite) TestNewUser_EmptyEmail() {
	_, err := NewUser("", "password123")

	s.Error(err)
	s.ErrorIs(err, ErrInvalidEmail)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(ErrCodeInvalidEmail, oopsErr.Code())
}

// Note: Current implementation only validates non-empty email
// More comprehensive email validation (regex) is not implemented
func (s *UserTestSuite) TestNewUser_ValidEmail() {
	testCases := []string{
		"user@example.com",
		"test.user@domain.co.uk",
		"user+tag@example.org",
		"123@numbers.com",
		"notanemail", // Currently passes - no format validation
	}

	for _, email := range testCases {
		user, err := NewUser(email, "password123")
		s.NoError(err, "Should succeed for email: %s", email)
		s.NotNil(user)
	}
}

func (s *UserTestSuite) TestNewUser_EmptyPassword() {
	_, err := NewUser("test@example.com", "")

	s.Error(err)
	s.ErrorIs(err, ErrPasswordTooShort)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(ErrCodePasswordTooShort, oopsErr.Code())
}

func (s *UserTestSuite) TestNewUser_ShortPassword() {
	_, err := NewUser("test@example.com", "short")

	s.Error(err)
	s.ErrorIs(err, ErrPasswordTooShort)
}

func (s *UserTestSuite) TestNewUser_PasswordHashing() {
	email := "test@example.com"
	password := "testPassword123"

	user1, err1 := NewUser(email, password)
	user2, err2 := NewUser(email, password)

	s.NoError(err1)
	s.NoError(err2)

	// Same password should produce different hashes (bcrypt uses salt)
	s.NotEqual(user1.password, user2.password)
}

// Test IsPasswordCorrect
func (s *UserTestSuite) TestIsPasswordCorrect_Success() {
	password := "correctPassword123"
	user, err := NewUser("test@example.com", password)
	s.NoError(err)

	isCorrect := user.IsPasswordCorrect(password)
	s.True(isCorrect)
}

func (s *UserTestSuite) TestIsPasswordCorrect_WrongPassword() {
	user, err := NewUser("test@example.com", "correctPassword123")
	s.NoError(err)

	isCorrect := user.IsPasswordCorrect("wrongPassword456")
	s.False(isCorrect)
}

func (s *UserTestSuite) TestIsPasswordCorrect_EmptyPassword() {
	user, err := NewUser("test@example.com", "password123")
	s.NoError(err)

	isCorrect := user.IsPasswordCorrect("")
	s.False(isCorrect)
}

// Test ReconstructUser
func (s *UserTestSuite) TestReconstructUser_Success() {
	id := int64(123)
	email := "test@example.com"
	passwordHash := "$2a$10$hashedpassword"
	createdAt := time.Now().Add(-24 * time.Hour)
	updatedAt := time.Now()

	user := ReconstructUser(id, email, passwordHash, createdAt, updatedAt)

	s.NotNil(user)
	s.Equal(id, user.ID)
	s.Equal(email, user.Email)
	s.Equal(passwordHash, user.password)
	s.Equal(createdAt, user.CreatedAt)
	s.Equal(updatedAt, user.UpdatedAt)
}

func (s *UserTestSuite) TestReconstructUser_PasswordAlreadyHashed() {
	passwordHash := "$2a$10$N9qo8uLOickgx2ZMRZoMye"
	user := ReconstructUser(1, "test@example.com", passwordHash, time.Now(), time.Now())

	s.Equal(passwordHash, user.password, "Password should not be re-hashed")
}

// Test ChangePassword
func (s *UserTestSuite) TestChangePassword_Success() {
	user, err := NewUser("test@example.com", "oldPassword123")
	s.NoError(err)

	err = user.ChangePassword("oldPassword123", "newPassword456")
	s.NoError(err)

	// Old password should no longer work
	s.False(user.IsPasswordCorrect("oldPassword123"))

	// New password should work
	s.True(user.IsPasswordCorrect("newPassword456"))
}

func (s *UserTestSuite) TestChangePassword_WrongOldPassword() {
	user, err := NewUser("test@example.com", "oldPassword123")
	s.NoError(err)

	err = user.ChangePassword("wrongOldPassword", "newPassword456")
	s.Error(err)
	s.ErrorIs(err, ErrIncorrectPassword)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(ErrCodeIncorrectPassword, oopsErr.Code())

	// Original password should still work
	s.True(user.IsPasswordCorrect("oldPassword123"))
}

func (s *UserTestSuite) TestChangePassword_ShortNewPassword() {
	user, err := NewUser("test@example.com", "oldPassword123")
	s.NoError(err)

	err = user.ChangePassword("oldPassword123", "short")
	s.Error(err)
	s.ErrorIs(err, ErrPasswordTooShort)

	// Original password should still work
	s.True(user.IsPasswordCorrect("oldPassword123"))
}

// Test business logic edge cases
func (s *UserTestSuite) TestUser_ImmutableID() {
	user, _ := NewUser("test@example.com", "password123")
	originalID := user.ID

	// Simulate repository assigning ID
	user.ID = 999

	s.Equal(int64(999), user.ID)
	s.NotEqual(originalID, user.ID)
}

func (s *UserTestSuite) TestUser_TimestampConsistency() {
	user, err := NewUser("test@example.com", "password123")
	s.NoError(err)

	// CreatedAt and UpdatedAt should be set to the same time on creation
	s.WithinDuration(user.CreatedAt, user.UpdatedAt, time.Second)
}
