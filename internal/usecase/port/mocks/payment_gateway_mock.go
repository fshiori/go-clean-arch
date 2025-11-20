package mocks

import (
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/stretchr/testify/mock"
)

// MockPaymentGateway is a mock implementation of port.PaymentGateway
type MockPaymentGateway struct {
	mock.Mock
}

// Ensure MockPaymentGateway implements port.PaymentGateway
var _ port.PaymentGateway = (*MockPaymentGateway)(nil)

// CreateTransaction mocks the CreateTransaction method
func (m *MockPaymentGateway) CreateTransaction(order *domain.Order, paymentInfo *port.PaymentInfo) (*port.Transaction, error) {
	args := m.Called(order, paymentInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.Transaction), args.Error(1)
}

// GetTransactionStatus mocks the GetTransactionStatus method
func (m *MockPaymentGateway) GetTransactionStatus(transactionID string) (port.TransactionStatus, error) {
	args := m.Called(transactionID)
	return args.Get(0).(port.TransactionStatus), args.Error(1)
}

// RefundTransaction mocks the RefundTransaction method
func (m *MockPaymentGateway) RefundTransaction(transactionID string, amount float64) error {
	args := m.Called(transactionID, amount)
	return args.Error(0)
}
