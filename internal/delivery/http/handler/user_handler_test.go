package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/mocks"

	"github.com/gin-gonic/gin"
	"github.com/samber/oops"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserHandlerTestSuite struct {
	suite.Suite
	router      *gin.Engine
	userUsecase *mocks.MockUserUsecase
	handler     *UserHandler
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (s *UserHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	s.userUsecase = new(mocks.MockUserUsecase)
	s.handler = NewUserHandler(s.userUsecase)

	s.router = gin.New()
	s.router.POST("/users", s.handler.CreateUser)
	s.router.GET("/users/:id", s.handler.GetUser)
	s.router.PUT("/users/:id/password", s.handler.UpdatePassword)
	s.router.GET("/users", s.handler.ListUsers)
	s.router.DELETE("/users/:id", s.handler.DeleteUser)
}

func (s *UserHandlerTestSuite) TearDownTest() {
	s.userUsecase.AssertExpectations(s.T())
}

// Helper to create test user
func (s *UserHandlerTestSuite) createTestUser() *domain.User {
	return domain.ReconstructUser(123, "test@example.com", "hashedpass", time.Now(), time.Now())
}

// Test CreateUser
func (s *UserHandlerTestSuite) TestCreateUser_Success() {
	reqBody := CreateUserRequest{
		Email:    "newuser@example.com",
		Password: "securePassword123",
	}

	expectedUser := domain.ReconstructUser(123, reqBody.Email, "hashedpass", time.Now(), time.Now())

	// Mock
	s.userUsecase.On("CreateUser", mock.Anything, reqBody.Email, reqBody.Password).Return(expectedUser, nil)

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusCreated, w.Code)

	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(int64(123), response.ID)
	s.Equal(reqBody.Email, response.Email)
}

func (s *UserHandlerTestSuite) TestCreateUser_InvalidJSON() {
	// Request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *UserHandlerTestSuite) TestCreateUser_ValidationFailed() {
	// Empty email
	reqBody := CreateUserRequest{
		Email:    "",
		Password: "password123",
	}

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *UserHandlerTestSuite) TestCreateUser_EmailAlreadyExists() {
	reqBody := CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	// Mock: Email already exists
	err := oops.Code(domain.ErrCodeEmailAlreadyExists).Wrap(domain.ErrEmailAlreadyExists)
	s.userUsecase.On("CreateUser", mock.Anything, reqBody.Email, reqBody.Password).Return(nil, err)

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusConflict, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal(domain.ErrCodeEmailAlreadyExists, response["code"])
}

// Test GetUser
func (s *UserHandlerTestSuite) TestGetUser_Success() {
	userID := int64(123)
	expectedUser := s.createTestUser()

	// Mock
	s.userUsecase.On("GetUserByID", mock.Anything, userID).Return(expectedUser, nil)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal(userID, response.ID)
	s.Equal("test@example.com", response.Email)
}

func (s *UserHandlerTestSuite) TestGetUser_NotFound() {
	userID := int64(999)

	// Mock
	err := oops.Code(domain.ErrCodeUserNotFound).Wrap(domain.ErrUserNotFound)
	s.userUsecase.On("GetUserByID", mock.Anything, userID).Return(nil, err)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *UserHandlerTestSuite) TestGetUser_InvalidID() {
	// Request with invalid ID
	req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Contains(response["error"], "invalid user ID format")
}

// Test UpdatePassword
func (s *UserHandlerTestSuite) TestUpdatePassword_Success() {
	userID := int64(123)
	reqBody := UpdatePasswordRequest{
		OldPassword: "oldPassword123",
		NewPassword: "newPassword456",
	}

	// Mock
	s.userUsecase.On("UpdateUserPassword", mock.Anything, userID, reqBody.OldPassword, reqBody.NewPassword).Return(nil)

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/123/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	s.Equal("password updated successfully", response["message"])
}

func (s *UserHandlerTestSuite) TestUpdatePassword_WrongOldPassword() {
	userID := int64(123)
	reqBody := UpdatePasswordRequest{
		OldPassword: "wrongOldPassword",
		NewPassword: "newPassword456",
	}

	// Mock
	err := oops.Code(domain.ErrCodeIncorrectPassword).Wrap(domain.ErrIncorrectPassword)
	s.userUsecase.On("UpdateUserPassword", mock.Anything, userID, reqBody.OldPassword, reqBody.NewPassword).Return(err)

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/123/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *UserHandlerTestSuite) TestUpdatePassword_InvalidID() {
	reqBody := UpdatePasswordRequest{
		OldPassword: "oldPass",
		NewPassword: "newPass",
	}

	// Request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/invalid/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)
}

// Test ListUsers
func (s *UserHandlerTestSuite) TestListUsers_Success() {
	user1 := domain.ReconstructUser(1, "user1@example.com", "hash1", time.Now(), time.Now())
	user2 := domain.ReconstructUser(2, "user2@example.com", "hash2", time.Now(), time.Now())
	expectedUsers := []*domain.User{user1, user2}

	// Mock
	s.userUsecase.On("ListUsers", mock.Anything, 1, 10).Return(expectedUsers, nil)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusOK, w.Code)

	var response []UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 2)
	s.Equal("user1@example.com", response[0].Email)
	s.Equal("user2@example.com", response[1].Email)
}

func (s *UserHandlerTestSuite) TestListUsers_DefaultPagination() {
	expectedUsers := []*domain.User{}

	// Mock: Default page=1, page_size=10
	s.userUsecase.On("ListUsers", mock.Anything, 1, 10).Return(expectedUsers, nil)

	// Request without query params
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusOK, w.Code)
}

// Test DeleteUser
func (s *UserHandlerTestSuite) TestDeleteUser_Success() {
	userID := int64(123)

	// Mock
	s.userUsecase.On("DeleteUser", mock.Anything, userID).Return(nil)

	// Request
	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusNoContent, w.Code)
}

func (s *UserHandlerTestSuite) TestDeleteUser_NotFound() {
	userID := int64(999)

	// Mock
	err := oops.Code(domain.ErrCodeUserNotFound).Wrap(domain.ErrUserNotFound)
	s.userUsecase.On("DeleteUser", mock.Anything, userID).Return(err)

	// Request
	req := httptest.NewRequest(http.MethodDelete, "/users/999", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *UserHandlerTestSuite) TestDeleteUser_InvalidID() {
	// Request
	req := httptest.NewRequest(http.MethodDelete, "/users/invalid", nil)
	w := httptest.NewRecorder()

	// Execute
	s.router.ServeHTTP(w, req)

	// Assert
	s.Equal(http.StatusBadRequest, w.Code)
}
