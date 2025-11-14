package handler

import (
	"net/http"
	"strconv"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase"
	"go-clean-arch/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userInteractor usecase.UserUsecase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userInteractor usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userInteractor: userInteractor,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Call usecase
	user, err := h.userInteractor.CreateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entity to response DTO
	response := toUserResponse(user)

	c.JSON(http.StatusCreated, response)
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Call usecase
	user, err := h.userInteractor.GetUserByID(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entity to response DTO
	response := toUserResponse(user)

	c.JSON(http.StatusOK, response)
}

// UpdatePassword handles PUT /users/:id/password
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Call usecase
	if err := h.userInteractor.UpdateUserPassword(c.Request.Context(), id, req.OldPassword, req.NewPassword); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// Call usecase
	users, err := h.userInteractor.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	// Convert domain entities to response DTOs
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = toUserResponse(user)
	}

	response := ListUsersResponse{
		Users:      userResponses,
		TotalCount: len(userResponses),
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Parse ID from URL
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Call usecase
	if err := h.userInteractor.DeleteUser(c.Request.Context(), id); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// toUserResponse converts a domain.User to a UserResponse DTO
// This is the critical transformation that happens at the delivery layer
func toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
