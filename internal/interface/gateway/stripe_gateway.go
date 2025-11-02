package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-clean-arch/internal/domain"
	"go-clean-arch/internal/usecase/port"
	"net/http"
	"time"
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
	Amount      int64  `json:"amount"`       // Amount in cents
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Source      string `json:"source"`       // Token or card ID
}

// stripeChargeResponse represents the response from Stripe API
type stripeChargeResponse struct {
	ID            string `json:"id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	FailureCode   string `json:"failure_code,omitempty"`
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
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", g.baseURL+"/charges", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var chargeResp stripeChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stripe error: %s - %s", chargeResp.FailureCode, chargeResp.FailureMessage)
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
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var chargeResp stripeChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
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
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", g.baseURL+"/refunds", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to process refund")
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
