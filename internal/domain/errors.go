package domain

import (
	"errors"
	"net/http"
)

// ============================================
// 領域錯誤碼（業務語義）
// ============================================

const (
	// User errors
	ErrCodeUserNotFound       = "USER_NOT_FOUND"
	ErrCodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	ErrCodeInvalidEmail       = "INVALID_EMAIL"
	ErrCodeInvalidPassword    = "INVALID_PASSWORD"
	ErrCodePasswordTooShort   = "PASSWORD_TOO_SHORT"
	ErrCodeIncorrectPassword  = "INCORRECT_PASSWORD"
	ErrCodePasswordHashFailed = "PASSWORD_HASH_FAILED"

	// Order errors
	ErrCodeOrderNotFound      = "ORDER_NOT_FOUND"
	ErrCodeEmptyOrder         = "EMPTY_ORDER"
	ErrCodeInvalidOrderStatus = "INVALID_ORDER_STATUS"
	ErrCodeInvalidUserID      = "INVALID_USER_ID"
	ErrCodeInvalidQuantity    = "INVALID_QUANTITY"

	// Payment errors
	ErrCodePaymentFailed     = "PAYMENT_FAILED"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"

	// Database errors
	ErrCodeDatabaseError       = "DATABASE_ERROR"
	ErrCodeDatabaseInsertFailed = "DATABASE_INSERT_FAILED"
	ErrCodeDatabaseUpdateFailed = "DATABASE_UPDATE_FAILED"
	ErrCodeDatabaseDeleteFailed = "DATABASE_DELETE_FAILED"
	ErrCodeDatabaseQueryFailed  = "DATABASE_QUERY_FAILED"

	// Repository errors
	ErrCodeUserSaveFailed   = "USER_SAVE_FAILED"
	ErrCodeUserUpdateFailed = "USER_UPDATE_FAILED"
	ErrCodeUserDeleteFailed = "USER_DELETE_FAILED"
	ErrCodeUserFetchFailed  = "USER_FETCH_FAILED"
	ErrCodeUserListFailed   = "USER_LIST_FAILED"

	ErrCodeOrderSaveFailed   = "ORDER_SAVE_FAILED"
	ErrCodeOrderUpdateFailed = "ORDER_UPDATE_FAILED"
	ErrCodeOrderFetchFailed  = "ORDER_FETCH_FAILED"
)

// ============================================
// 標準領域錯誤（可用 errors.Is 判斷）
// ============================================

var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrIncorrectPassword  = errors.New("incorrect password")
	ErrPasswordHashFailed = errors.New("failed to hash password")

	// Order errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrEmptyOrder         = errors.New("order must have at least one item")
	ErrInvalidOrderStatus = errors.New("invalid order status transition")
	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidQuantity    = errors.New("invalid quantity")

	// Payment errors
	ErrPaymentFailed = errors.New("payment failed")
)

// ============================================
// 錯誤碼到 HTTP 狀態碼的映射
// ============================================

var ErrorCodeToHTTPStatus = map[string]int{
	// 404 Not Found
	ErrCodeUserNotFound:  http.StatusNotFound,
	ErrCodeOrderNotFound: http.StatusNotFound,

	// 409 Conflict
	ErrCodeEmailAlreadyExists: http.StatusConflict,

	// 400 Bad Request
	ErrCodeInvalidEmail:       http.StatusBadRequest,
	ErrCodeInvalidPassword:    http.StatusBadRequest,
	ErrCodePasswordTooShort:   http.StatusBadRequest,
	ErrCodeIncorrectPassword:  http.StatusBadRequest,
	ErrCodeEmptyOrder:         http.StatusBadRequest,
	ErrCodeInvalidOrderStatus: http.StatusBadRequest,
	ErrCodeInvalidUserID:      http.StatusBadRequest,
	ErrCodeInvalidQuantity:    http.StatusBadRequest,

	// 402 Payment Required
	ErrCodePaymentFailed:     http.StatusPaymentRequired,
	ErrCodeInsufficientFunds: http.StatusPaymentRequired,

	// 500 Internal Server Error
	ErrCodePasswordHashFailed:   http.StatusInternalServerError,
	ErrCodeDatabaseError:        http.StatusInternalServerError,
	ErrCodeDatabaseInsertFailed: http.StatusInternalServerError,
	ErrCodeDatabaseUpdateFailed: http.StatusInternalServerError,
	ErrCodeDatabaseDeleteFailed: http.StatusInternalServerError,
	ErrCodeDatabaseQueryFailed:  http.StatusInternalServerError,
	ErrCodeUserSaveFailed:       http.StatusInternalServerError,
	ErrCodeUserUpdateFailed:     http.StatusInternalServerError,
	ErrCodeUserDeleteFailed:     http.StatusInternalServerError,
	ErrCodeUserFetchFailed:      http.StatusInternalServerError,
	ErrCodeUserListFailed:       http.StatusInternalServerError,
	ErrCodeOrderSaveFailed:      http.StatusInternalServerError,
	ErrCodeOrderUpdateFailed:    http.StatusInternalServerError,
	ErrCodeOrderFetchFailed:     http.StatusInternalServerError,
}
