package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"go-clean-arch/internal/usecase"
	"go-clean-arch/pkg/logger"

	"github.com/google/uuid"
)

// OrderConsumer consumes order-related messages from a message queue
type OrderConsumer struct {
	orderInteractor usecase.OrderUsecase
}

// NewOrderConsumer creates a new OrderConsumer
func NewOrderConsumer(orderInteractor usecase.OrderUsecase) *OrderConsumer {
	return &OrderConsumer{
		orderInteractor: orderInteractor,
	}
}

// OrderMessage represents a message from the queue
type OrderMessage struct {
	Type    string          `json:"type"`
	OrderID int64           `json:"order_id"`
	Payload json.RawMessage `json:"payload"`
}

// ConsumeMessage processes a message from the queue
func (c *OrderConsumer) ConsumeMessage(messageBody []byte) error {
	// Create context with trace ID for this message
	ctx := context.Background()
	traceID := uuid.New().String()
	ctx = logger.WithTraceID(ctx, traceID)

	var msg OrderMessage
	if err := json.Unmarshal(messageBody, &msg); err != nil {
		logger.ErrorContext(ctx, "Failed to unmarshal message", "error", err)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	logger.InfoContext(ctx, "Processing message",
		"type", msg.Type,
		"order_id", msg.OrderID,
	)

	var err error
	switch msg.Type {
	case "order.ship":
		err = c.handleShipOrder(ctx, msg.OrderID)
	case "order.complete":
		err = c.handleCompleteOrder(ctx, msg.OrderID)
	case "order.cancel":
		err = c.handleCancelOrder(ctx, msg.OrderID)
	default:
		err = fmt.Errorf("unknown message type: %s", msg.Type)
		logger.ErrorContext(ctx, "Unknown message type", "type", msg.Type)
	}

	if err != nil {
		logger.ErrorContext(ctx, "Failed to process message",
			"type", msg.Type,
			"order_id", msg.OrderID,
			"error", err,
		)
	}

	return err
}

// handleShipOrder processes order shipment
func (c *OrderConsumer) handleShipOrder(ctx context.Context, orderID int64) error {
	logger.InfoContext(ctx, "Shipping order", "order_id", orderID)

	if err := c.orderInteractor.ShipOrder(ctx, orderID); err != nil {
		logger.ErrorContext(ctx, "Failed to ship order", "order_id", orderID, "error", err)
		return fmt.Errorf("failed to ship order: %w", err)
	}

	logger.InfoContext(ctx, "Order shipped successfully", "order_id", orderID)
	return nil
}

// handleCompleteOrder processes order completion
func (c *OrderConsumer) handleCompleteOrder(ctx context.Context, orderID int64) error {
	logger.InfoContext(ctx, "Completing order", "order_id", orderID)

	if err := c.orderInteractor.CompleteOrder(ctx, orderID); err != nil {
		logger.ErrorContext(ctx, "Failed to complete order", "order_id", orderID, "error", err)
		return fmt.Errorf("failed to complete order: %w", err)
	}

	logger.InfoContext(ctx, "Order completed successfully", "order_id", orderID)
	return nil
}

// handleCancelOrder processes order cancellation
func (c *OrderConsumer) handleCancelOrder(ctx context.Context, orderID int64) error {
	logger.InfoContext(ctx, "Cancelling order", "order_id", orderID)

	if err := c.orderInteractor.CancelOrder(ctx, orderID); err != nil {
		logger.ErrorContext(ctx, "Failed to cancel order", "order_id", orderID, "error", err)
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	logger.InfoContext(ctx, "Order cancelled successfully", "order_id", orderID)
	return nil
}

// Start starts consuming messages from the queue
// This is a simplified example; in production, you'd connect to a real message queue
func (c *OrderConsumer) Start(ctx context.Context) error {
	logger.Info("Starting order consumer...")

	// TODO: Connect to message queue (RabbitMQ, Kafka, etc.)
	// Example:
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	// if err != nil {
	//     logger.Error("Failed to connect to RabbitMQ", "error", err)
	//     return err
	// }
	// defer conn.Close()
	//
	// ch, err := conn.Channel()
	// if err != nil {
	//     logger.Error("Failed to open channel", "error", err)
	//     return err
	// }
	// defer ch.Close()
	//
	// msgs, err := ch.Consume(
	//     "order_queue", // queue name
	//     "",            // consumer
	//     false,         // auto-ack (set to false for graceful shutdown)
	//     false,         // exclusive
	//     false,         // no-local
	//     false,         // no-wait
	//     nil,           // args
	// )
	// if err != nil {
	//     logger.Error("Failed to register consumer", "error", err)
	//     return err
	// }
	//
	// logger.Info("Order consumer connected and waiting for messages...")
	//
	// // Consume messages with graceful shutdown support
	// for {
	//     select {
	//     case msg, ok := <-msgs:
	//         if !ok {
	//             logger.Info("Message channel closed")
	//             return nil
	//         }
	//         if err := c.ConsumeMessage(msg.Body); err != nil {
	//             logger.Error("Error processing message", "error", err)
	//             msg.Nack(false, true) // Negative ack, requeue
	//         } else {
	//             msg.Ack(false) // Acknowledge successful processing
	//         }
	//     case <-ctx.Done():
	//         logger.Info("Shutdown signal received, stopping consumer...")
	//         return nil
	//     }
	// }

	logger.Info("Order consumer is running...")

	// Wait for context cancellation (graceful shutdown)
	<-ctx.Done()
	logger.Info("Consumer shutdown signal received")
	return nil
}
