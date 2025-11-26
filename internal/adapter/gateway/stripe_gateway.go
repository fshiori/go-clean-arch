package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"

	"github.com/samber/oops"
)

// stripeGateway implements the PaymentGateway interface for Stripe
type stripeGateway struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewStripeGateway creates a new Stripe payment gateway
func NewStripeGateway(apiKey string) port.PaymentGateway {
	return &stripeGateway{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "https://api.stripe.com/v1",
	}
}

// stripeChargeRequest represents the request to Stripe API
type stripeChargeRequest struct {
	Amount      int64  `json:"amount"` // Amount in cents
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Source      string `json:"source"` // Token or card ID
}

// stripeChargeResponse represents the response from Stripe API
type stripeChargeResponse struct {
	ID             string `json:"id"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"`
	FailureCode    string `json:"failure_code,omitempty"`
	FailureMessage string `json:"failure_message,omitempty"`
}

// CreateTransaction creates a new payment transaction using Stripe
func (g *stripeGateway) CreateTransaction(order *domain.Order, paymentInfo *port.PaymentInfo) (*port.Transaction, error) {
	// Convert order to Stripe request
	chargeReq := stripeChargeRequest{
		Amount:      int64(order.TotalAmount * 100), // Convert to cents
		Currency:    "usd",
		Description: fmt.Sprintf("Payment for Order #%d", order.ID),
		Source:      paymentInfo.CardNumber, // In real implementation, this would be a token
	}

	// Marshal request
	reqBody, err := json.Marshal(chargeReq)
	if err != nil {
		return nil, oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "marshal").
			With("order_id", order.ID).
			Hint("Failed to marshal payment request").
			Wrapf(err, "failed to marshal Stripe request")
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", g.baseURL+"/charges", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http").
			With("order_id", order.ID).
			Wrapf(err, "failed to create HTTP request")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http", "network").
			With("order_id", order.ID).
			With("url", g.baseURL+"/charges").
			Hint("Check network connectivity and Stripe API status").
			Wrapf(err, "failed to execute Stripe API request")
	}
	defer resp.Body.Close()

	// Parse response
	var chargeResp stripeChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return nil, oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "decode").
			With("order_id", order.ID).
			With("status_code", resp.StatusCode).
			Wrapf(err, "failed to decode Stripe response")
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, oops.
			Code("PAYMENT_DECLINED").
			In("gateway").
			Tags("stripe", "declined").
			With("order_id", order.ID).
			With("failure_code", chargeResp.FailureCode).
			With("status_code", resp.StatusCode).
			Hint(chargeResp.FailureMessage).
			Errorf("payment declined by Stripe")
	}

	// Convert Stripe response to domain Transaction
	transaction := &port.Transaction{
		ID:            chargeResp.ID,
		OrderID:       order.ID,
		Amount:        float64(chargeResp.Amount) / 100, // Convert back from cents
		Status:        mapStripeStatus(chargeResp.Status),
		PaymentMethod: "stripe",
		TransactionData: map[string]interface{}{
			"stripe_charge_id": chargeResp.ID,
			"stripe_status":    chargeResp.Status,
		},
	}

	return transaction, nil
}

// GetTransactionStatus retrieves the status of a transaction
func (g *stripeGateway) GetTransactionStatus(transactionID string) (port.TransactionStatus, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", g.baseURL+"/charges/"+transactionID, nil)
	if err != nil {
		return "", oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http").
			With("transaction_id", transactionID).
			Wrapf(err, "failed to create HTTP request")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http", "network").
			With("transaction_id", transactionID).
			Hint("Check network connectivity and Stripe API status").
			Wrapf(err, "failed to execute Stripe API request")
	}
	defer resp.Body.Close()

	// Parse response
	var chargeResp stripeChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return "", oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "decode").
			With("transaction_id", transactionID).
			With("status_code", resp.StatusCode).
			Wrapf(err, "failed to decode Stripe response")
	}

	return mapStripeStatus(chargeResp.Status), nil
}

// RefundTransaction refunds a completed transaction
func (g *stripeGateway) RefundTransaction(transactionID string, amount float64) error {
	// Prepare refund request
	refundReq := map[string]interface{}{
		"charge": transactionID,
	}

	if amount > 0 {
		refundReq["amount"] = int64(amount * 100) // Convert to cents
	}

	// Marshal request
	reqBody, err := json.Marshal(refundReq)
	if err != nil {
		return oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "marshal").
			With("transaction_id", transactionID).
			With("amount", amount).
			Wrapf(err, "failed to marshal refund request")
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", g.baseURL+"/refunds", bytes.NewBuffer(reqBody))
	if err != nil {
		return oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http").
			With("transaction_id", transactionID).
			Wrapf(err, "failed to create HTTP request")
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return oops.
			Code("PAYMENT_GATEWAY_ERROR").
			In("gateway").
			Tags("stripe", "http", "network").
			With("transaction_id", transactionID).
			Hint("Check network connectivity and Stripe API status").
			Wrapf(err, "failed to execute Stripe refund request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return oops.
			Code("REFUND_FAILED").
			In("gateway").
			Tags("stripe", "refund").
			With("transaction_id", transactionID).
			With("status_code", resp.StatusCode).
			Hint("Refund was rejected by Stripe").
			Errorf("failed to process refund")
	}

	return nil
}

// mapStripeStatus maps Stripe status to our domain TransactionStatus
func mapStripeStatus(stripeStatus string) port.TransactionStatus {
	switch stripeStatus {
	case "succeeded":
		return port.TransactionStatusCompleted
	case "pending":
		return port.TransactionStatusPending
	case "failed":
		return port.TransactionStatusFailed
	default:
		return port.TransactionStatusFailed
	}
}
