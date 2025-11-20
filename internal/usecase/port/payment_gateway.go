package port

import "go-clean-arch/internal/domain"

// TransactionStatus represents the status of a payment transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID              string
	OrderID         int64
	Amount          float64
	Status          TransactionStatus
	PaymentMethod   string
	TransactionData map[string]interface{}
}

// PaymentInfo contains payment information
type PaymentInfo struct {
	CardNumber string
	CVV        string
	ExpiryDate string
	// Add other payment-related fields as needed
}

// PaymentGateway defines the interface for payment processing
// This abstracts away the specific payment provider (Stripe, PayPal, etc.)
type PaymentGateway interface {
	// CreateTransaction creates a new payment transaction
	CreateTransaction(order *domain.Order, paymentInfo *PaymentInfo) (*Transaction, error)

	// GetTransactionStatus retrieves the status of a transaction
	GetTransactionStatus(transactionID string) (TransactionStatus, error)

	// RefundTransaction refunds a completed transaction
	RefundTransaction(transactionID string, amount float64) error
}
