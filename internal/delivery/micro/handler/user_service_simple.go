// Package handler provides go-micro service handlers.
// This delivery layer uses JSON encoding for simplicity.
package handler

import (
	"context"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase"

	"github.com/samber/oops"
)

// UserServiceSimple implements a simple JSON-based user service
type UserServiceSimple struct {
	userUsecase usecase.UserUsecase
}

// NewUserServiceSimple creates a new UserServiceSimple handler
func NewUserServiceSimple(userUsecase usecase.UserUsecase) *UserServiceSimple {
	return &UserServiceSimple{
		userUsecase: userUsecase,
	}
}

// CreateUserReq represents the create user request
type CreateUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUserRsp represents the create user response
type CreateUserRsp struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateUser creates a new user
func (s *UserServiceSimple) CreateUser(ctx context.Context, req *CreateUserReq, rsp *CreateUserRsp) error {
	// Call usecase
	user, err := s.userUsecase.CreateUser(ctx, req.Email, req.Password)
	if err != nil {
		return convertDomainError(err)
	}

	// Convert domain entity to response
	rsp.ID = user.ID
	rsp.Email = user.Email
	rsp.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	rsp.UpdatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return nil
}

// GetUserReq represents the get user request
type GetUserReq struct {
	ID int64 `json:"id"`
}

// GetUserRsp represents the get user response
type GetUserRsp struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetUser retrieves a user by ID
func (s *UserServiceSimple) GetUser(ctx context.Context, req *GetUserReq, rsp *GetUserRsp) error {
	// Call usecase
	user, err := s.userUsecase.GetUserByID(ctx, req.ID)
	if err != nil {
		return convertDomainError(err)
	}

	// Convert domain entity to response
	rsp.ID = user.ID
	rsp.Email = user.Email
	rsp.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	rsp.UpdatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return nil
}

// ListUsersReq represents the list users request
type ListUsersReq struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// UserInfo represents user information in the response
type UserInfo struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListUsersRsp represents the list users response
type ListUsersRsp struct {
	Users      []*UserInfo `json:"users"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

// ListUsers lists users with pagination
func (s *UserServiceSimple) ListUsers(ctx context.Context, req *ListUsersReq, rsp *ListUsersRsp) error {
	// Default values
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}

	// Call usecase
	users, err := s.userUsecase.ListUsers(ctx, page, pageSize)
	if err != nil {
		return convertDomainError(err)
	}

	// Convert domain entities to response
	rsp.Users = make([]*UserInfo, len(users))
	for i, user := range users {
		rsp.Users[i] = &UserInfo{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	rsp.TotalCount = len(users)
	rsp.Page = page
	rsp.PageSize = pageSize

	return nil
}

// UpdatePasswordReq represents the update password request
type UpdatePasswordReq struct {
	ID          int64  `json:"id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// UpdatePasswordRsp represents the update password response
type UpdatePasswordRsp struct {
	Message string `json:"message"`
}

// UpdatePassword updates a user's password
func (s *UserServiceSimple) UpdatePassword(ctx context.Context, req *UpdatePasswordReq, rsp *UpdatePasswordRsp) error {
	// Call usecase
	if err := s.userUsecase.UpdateUserPassword(ctx, req.ID, req.OldPassword, req.NewPassword); err != nil {
		return convertDomainError(err)
	}

	rsp.Message = "password updated successfully"
	return nil
}

// DeleteUserReq represents the delete user request
type DeleteUserReq struct {
	ID int64 `json:"id"`
}

// DeleteUserRsp represents the delete user response
type DeleteUserRsp struct {
	Message string `json:"message"`
}

// DeleteUser deletes a user
func (s *UserServiceSimple) DeleteUser(ctx context.Context, req *DeleteUserReq, rsp *DeleteUserRsp) error {
	// Call usecase
	if err := s.userUsecase.DeleteUser(ctx, req.ID); err != nil {
		return convertDomainError(err)
	}

	rsp.Message = "user deleted successfully"
	return nil
}

// convertDomainError converts domain errors to appropriate errors
func convertDomainError(err error) error {
	// Extract oops error information if available
	ooErr, ok := err.(oops.OopsError)
	if !ok {
		return err
	}

	// Check for domain-specific errors
	switch ooErr.Code() {
	case domain.ErrCodeUserNotFound:
		return oops.
			Code(domain.ErrCodeUserNotFound).
			With("message", "user not found").
			Errorf("user not found")
	case domain.ErrCodeUserAlreadyExists:
		return oops.
			Code(domain.ErrCodeUserAlreadyExists).
			With("message", "user already exists").
			Errorf("user already exists")
	case domain.ErrCodeInvalidPassword:
		return oops.
			Code(domain.ErrCodeInvalidPassword).
			With("message", "invalid password").
			Errorf("invalid password")
	case domain.ErrCodeValidationFailed:
		return oops.
			Code(domain.ErrCodeValidationFailed).
			With("message", "validation failed").
			Errorf("validation failed")
	default:
		return err
	}
}
