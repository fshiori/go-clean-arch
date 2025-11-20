// Package handler provides go-micro service handlers.
// This delivery layer converts between protobuf messages and domain entities.
package handler

import (
	"context"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase"
	pb "go-clean-arch/proto/user"

	"github.com/samber/oops"
)

// UserService implements the UserService proto interface
type UserService struct {
	userUsecase usecase.UserUsecase
}

// NewUserService creates a new UserService handler
func NewUserService(userUsecase usecase.UserUsecase) *UserService {
	return &UserService{
		userUsecase: userUsecase,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest, rsp *pb.CreateUserResponse) error {
	// Call usecase
	user, err := s.userUsecase.CreateUser(ctx, req.Email, req.Password)
	if err != nil {
		return convertError(err)
	}

	// Convert domain entity to proto response
	rsp.Id = user.ID
	rsp.Email = user.Email
	rsp.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	rsp.UpdatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest, rsp *pb.GetUserResponse) error {
	// Call usecase
	user, err := s.userUsecase.GetUserByID(ctx, req.Id)
	if err != nil {
		return convertError(err)
	}

	// Convert domain entity to proto response
	rsp.Id = user.ID
	rsp.Email = user.Email
	rsp.CreatedAt = user.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
	rsp.UpdatedAt = user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")

	return nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(ctx context.Context, req *pb.ListUsersRequest, rsp *pb.ListUsersResponse) error {
	// Default values
	page := int(req.Page)
	if page == 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize == 0 {
		pageSize = 10
	}

	// Call usecase
	users, err := s.userUsecase.ListUsers(ctx, page, pageSize)
	if err != nil {
		return convertError(err)
	}

	// Convert domain entities to proto response
	rsp.Users = make([]*pb.UserInfo, len(users))
	for i, user := range users {
		rsp.Users[i] = &pb.UserInfo{
			Id:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	rsp.TotalCount = int32(len(users))
	rsp.Page = int32(page)
	rsp.PageSize = int32(pageSize)

	return nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(ctx context.Context, req *pb.UpdatePasswordRequest, rsp *pb.UpdatePasswordResponse) error {
	// Call usecase
	if err := s.userUsecase.UpdateUserPassword(ctx, req.Id, req.OldPassword, req.NewPassword); err != nil {
		return convertError(err)
	}

	rsp.Message = "password updated successfully"
	return nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest, rsp *pb.DeleteUserResponse) error {
	// Call usecase
	if err := s.userUsecase.DeleteUser(ctx, req.Id); err != nil {
		return convertError(err)
	}

	rsp.Message = "user deleted successfully"
	return nil
}

// convertError converts domain errors to appropriate micro errors
func convertError(err error) error {
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
