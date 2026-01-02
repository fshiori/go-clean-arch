package usecase

import (
	"context"
	"testing"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port/mocks"

	"github.com/samber/oops"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	testEmail    = "user@example.com"
	testPassword = "password123"
)

type UserInteractorTestSuite struct {
	suite.Suite
	userRepo   *mocks.MockUserRepository
	interactor *UserInteractor
	ctx        context.Context
}

func TestUserInteractorTestSuite(t *testing.T) {
	suite.Run(t, new(UserInteractorTestSuite))
}

func (s *UserInteractorTestSuite) SetupTest() {
	s.userRepo = new(mocks.MockUserRepository)
	s.interactor = NewUserInteractor(s.userRepo)
	s.ctx = context.Background()
}

func (s *UserInteractorTestSuite) TearDownTest() {
	s.userRepo.AssertExpectations(s.T())
}

// Test CreateUser
func (s *UserInteractorTestSuite) TestCreateUser_Success() {
	email := "newuser@example.com"
	password := "securePassword123"

	// Mock: FindByEmail returns nil (user doesn't exist)
	s.userRepo.On("FindByEmail", s.ctx, email).Return(nil, domain.ErrUserNotFound)

	// Mock: Save succeeds and sets the ID
	s.userRepo.On("Save", s.ctx, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == email && u.ID == 0
	})).Run(func(args mock.Arguments) {
		// Simulate repository setting the ID
		user := args.Get(1).(*domain.User)
		user.ID = 123
	}).Return(nil)

	// Execute
	user, err := s.interactor.CreateUser(s.ctx, email, password)

	// Assert
	s.NoError(err)
	s.NotNil(user)
	s.Equal(email, user.Email)
	s.Equal(int64(123), user.ID)
}

func (s *UserInteractorTestSuite) TestCreateUser_EmailAlreadyExists() {
	email := "existing@example.com"
	password := testPassword

	existingUser := domain.ReconstructUser(1, email, "hashedpass", time.Now(), time.Now())

	// Mock: FindByEmail returns existing user
	s.userRepo.On("FindByEmail", s.ctx, email).Return(existingUser, nil)

	// Execute
	user, err := s.interactor.CreateUser(s.ctx, email, password)

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrEmailAlreadyExists)

	oopsErr, ok := oops.AsOops(err)
	s.True(ok)
	s.Equal(domain.ErrCodeEmailAlreadyExists, oopsErr.Code())
}

func (s *UserInteractorTestSuite) TestCreateUser_InvalidEmail() {
	email := ""
	password := testPassword

	// Mock: FindByEmail is called first (even for empty email)
	s.userRepo.On("FindByEmail", s.ctx, email).Return(nil, domain.ErrUserNotFound)

	// Execute
	user, err := s.interactor.CreateUser(s.ctx, email, password)

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrInvalidEmail)
}

func (s *UserInteractorTestSuite) TestCreateUser_InvalidPassword() {
	email := testEmail
	password := "short"

	// Mock: FindByEmail is called first
	s.userRepo.On("FindByEmail", s.ctx, email).Return(nil, domain.ErrUserNotFound)

	// Execute
	user, err := s.interactor.CreateUser(s.ctx, email, password)

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrPasswordTooShort)
}

func (s *UserInteractorTestSuite) TestCreateUser_RepositoryError() {
	email := testEmail
	password := testPassword

	// Mock: FindByEmail returns nil (no duplicate)
	s.userRepo.On("FindByEmail", s.ctx, email).Return(nil, domain.ErrUserNotFound)

	// Mock: Save fails with database error
	dbErr := oops.Code("DB_ERROR").Errorf("database connection failed")
	s.userRepo.On("Save", s.ctx, mock.AnythingOfType("*domain.User")).Return(dbErr)

	// Execute
	user, err := s.interactor.CreateUser(s.ctx, email, password)

	// Assert
	s.Error(err)
	s.Nil(user)
}

// Test GetUserByID
func (s *UserInteractorTestSuite) TestGetUserByID_Success() {
	userID := int64(123)
	expectedUser := domain.ReconstructUser(userID, "user@example.com", "hashedpass", time.Now(), time.Now())

	// Mock
	s.userRepo.On("FindByID", s.ctx, userID).Return(expectedUser, nil)

	// Execute
	user, err := s.interactor.GetUserByID(s.ctx, userID)

	// Assert
	s.NoError(err)
	s.NotNil(user)
	s.Equal(userID, user.ID)
	s.Equal("user@example.com", user.Email)
}

func (s *UserInteractorTestSuite) TestGetUserByID_NotFound() {
	userID := int64(999)

	// Mock
	s.userRepo.On("FindByID", s.ctx, userID).Return(nil, domain.ErrUserNotFound)

	// Execute
	user, err := s.interactor.GetUserByID(s.ctx, userID)

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserInteractorTestSuite) TestGetUserByID_InvalidID() {
	invalidIDs := []int64{0, -1, -100}

	for _, id := range invalidIDs {
		// No mock needed - validation happens first

		// Execute
		user, err := s.interactor.GetUserByID(s.ctx, id)

		// Assert
		s.Error(err, "Should fail for ID: %d", id)
		s.Nil(user)
		s.ErrorIs(err, domain.ErrInvalidUserID)
	}
}

// Test GetUserByEmail
func (s *UserInteractorTestSuite) TestGetUserByEmail_Success() {
	email := "user@example.com"
	expectedUser := domain.ReconstructUser(123, email, "hashedpass", time.Now(), time.Now())

	// Mock
	s.userRepo.On("FindByEmail", s.ctx, email).Return(expectedUser, nil)

	// Execute
	user, err := s.interactor.GetUserByEmail(s.ctx, email)

	// Assert
	s.NoError(err)
	s.NotNil(user)
	s.Equal(email, user.Email)
}

func (s *UserInteractorTestSuite) TestGetUserByEmail_NotFound() {
	email := "nonexistent@example.com"

	// Mock
	s.userRepo.On("FindByEmail", s.ctx, email).Return(nil, domain.ErrUserNotFound)

	// Execute
	user, err := s.interactor.GetUserByEmail(s.ctx, email)

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserInteractorTestSuite) TestGetUserByEmail_EmptyEmail() {
	// No mock needed

	// Execute
	user, err := s.interactor.GetUserByEmail(s.ctx, "")

	// Assert
	s.Error(err)
	s.Nil(user)
	s.ErrorIs(err, domain.ErrInvalidEmail)
}

// Test DeleteUser
func (s *UserInteractorTestSuite) TestDeleteUser_Success() {
	userID := int64(123)

	// Mock: Delete succeeds
	s.userRepo.On("Delete", s.ctx, userID).Return(nil)

	// Execute
	err := s.interactor.DeleteUser(s.ctx, userID)

	// Assert
	s.NoError(err)
}

func (s *UserInteractorTestSuite) TestDeleteUser_RepositoryError() {
	userID := int64(999)

	// Mock: Delete returns error
	deleteErr := oops.Code("DB_ERROR").Errorf("user not found")
	s.userRepo.On("Delete", s.ctx, userID).Return(deleteErr)

	// Execute
	err := s.interactor.DeleteUser(s.ctx, userID)

	// Assert
	s.Error(err)
}

func (s *UserInteractorTestSuite) TestDeleteUser_InvalidID() {
	// No mock needed

	// Execute
	err := s.interactor.DeleteUser(s.ctx, 0)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrInvalidUserID)
}

// Test ListUsers
func (s *UserInteractorTestSuite) TestListUsers_Success() {
	page := 1
	pageSize := 10

	expectedUsers := []*domain.User{
		domain.ReconstructUser(1, "user1@example.com", "hash1", time.Now(), time.Now()),
		domain.ReconstructUser(2, "user2@example.com", "hash2", time.Now(), time.Now()),
		domain.ReconstructUser(3, "user3@example.com", "hash3", time.Now(), time.Now()),
	}

	// Mock
	s.userRepo.On("List", s.ctx, 0, 10).Return(expectedUsers, nil)

	// Execute
	users, err := s.interactor.ListUsers(s.ctx, page, pageSize)

	// Assert
	s.NoError(err)
	s.Len(users, 3)
	s.Equal("user1@example.com", users[0].Email)
}

func (s *UserInteractorTestSuite) TestListUsers_InvalidPaginationAutoCorrects() {
	testCases := []struct {
		page           int
		pageSize       int
		expectedOffset int
		expectedLimit  int
		description    string
	}{
		{0, 10, 0, 10, "page < 1 becomes 1"},
		{-1, 10, 0, 10, "negative page becomes 1"},
		{1, 0, 0, 10, "pageSize < 1 becomes 10"},
		{1, -1, 0, 10, "negative pageSize becomes 10"},
		{1, 101, 0, 10, "pageSize > 100 becomes 10"},
		{2, 200, 10, 10, "page 2 with oversized pageSize"},
	}

	for _, tc := range testCases {
		// Mock with the corrected pagination
		s.userRepo.On("List", s.ctx, tc.expectedOffset, tc.expectedLimit).
			Return([]*domain.User{}, nil).Once()

		// Execute
		users, err := s.interactor.ListUsers(s.ctx, tc.page, tc.pageSize)

		// Assert
		s.NoError(err, "Should auto-correct for: %s", tc.description)
		s.NotNil(users, "Should auto-correct for: %s", tc.description)
	}
}

func (s *UserInteractorTestSuite) TestListUsers_EmptyResult() {
	page := 1
	pageSize := 10

	// Mock: Return empty list
	s.userRepo.On("List", s.ctx, 0, 10).Return([]*domain.User{}, nil)

	// Execute
	users, err := s.interactor.ListUsers(s.ctx, page, pageSize)

	// Assert
	s.NoError(err)
	s.Empty(users)
}

// Test UpdateUserPassword
func (s *UserInteractorTestSuite) TestUpdateUserPassword_Success() {
	userID := int64(123)
	oldPassword := "oldPassword123"
	newPassword := "newPassword456"

	// Create user with known password
	user, err := domain.NewUser("user@example.com", oldPassword)
	s.NoError(err)
	user.ID = userID

	// Mock: FindByID returns the user
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Mock: Update succeeds
	s.userRepo.On("Update", s.ctx, user).Return(nil)

	// Execute
	err = s.interactor.UpdateUserPassword(s.ctx, userID, oldPassword, newPassword)

	// Assert
	s.NoError(err)

	// Verify password was changed
	s.True(user.IsPasswordCorrect(newPassword))
	s.False(user.IsPasswordCorrect(oldPassword))
}

func (s *UserInteractorTestSuite) TestUpdateUserPassword_UserNotFound() {
	userID := int64(999)

	// Mock
	s.userRepo.On("FindByID", s.ctx, userID).Return(nil, domain.ErrUserNotFound)

	// Execute
	err := s.interactor.UpdateUserPassword(s.ctx, userID, "oldPass", "newPass")

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserNotFound)
}

func (s *UserInteractorTestSuite) TestUpdateUserPassword_WrongOldPassword() {
	userID := int64(123)
	correctPassword := "correctPassword123"
	wrongOldPassword := "wrongOldPassword"
	newPassword := "newPassword456"

	user, err := domain.NewUser("user@example.com", correctPassword)
	s.NoError(err)
	user.ID = userID

	// Mock
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Execute
	err = s.interactor.UpdateUserPassword(s.ctx, userID, wrongOldPassword, newPassword)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrIncorrectPassword)

	// Password should not have changed
	s.True(user.IsPasswordCorrect(correctPassword))
}

func (s *UserInteractorTestSuite) TestUpdateUserPassword_InvalidNewPassword() {
	userID := int64(123)
	oldPassword := "oldPassword123"
	shortPassword := "short"

	user, err := domain.NewUser("user@example.com", oldPassword)
	s.NoError(err)
	user.ID = userID

	// Mock
	s.userRepo.On("FindByID", s.ctx, userID).Return(user, nil)

	// Execute
	err = s.interactor.UpdateUserPassword(s.ctx, userID, oldPassword, shortPassword)

	// Assert
	s.Error(err)
	s.ErrorIs(err, domain.ErrPasswordTooShort)

	// Password should not have changed
	s.True(user.IsPasswordCorrect(oldPassword))
}
