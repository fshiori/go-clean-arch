package consumer

import (
	"encoding/json"
	"fmt"
	"go-clean-arch/internal/usecase"
	"log"
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
	var msg OrderMessage
	if err := json.Unmarshal(messageBody, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("Processing message type: %s for order: %d", msg.Type, msg.OrderID)

	switch msg.Type {
	case "order.ship":
		return c.handleShipOrder(msg.OrderID)
	case "order.complete":
		return c.handleCompleteOrder(msg.OrderID)
	case "order.cancel":
		return c.handleCancelOrder(msg.OrderID)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleShipOrder processes order shipment
func (c *OrderConsumer) handleShipOrder(orderID int64) error {
	log.Printf("Shipping order: %d", orderID)

	if err := c.orderInteractor.ShipOrder(orderID); err != nil {
		return fmt.Errorf("failed to ship order: %w", err)
	}

	log.Printf("Order %d shipped successfully", orderID)
	return nil
}

// handleCompleteOrder processes order completion
func (c *OrderConsumer) handleCompleteOrder(orderID int64) error {
	log.Printf("Completing order: %d", orderID)

	if err := c.orderInteractor.CompleteOrder(orderID); err != nil {
		return fmt.Errorf("failed to complete order: %w", err)
	}

	log.Printf("Order %d completed successfully", orderID)
	return nil
}

// handleCancelOrder processes order cancellation
func (c *OrderConsumer) handleCancelOrder(orderID int64) error {
	log.Printf("Cancelling order: %d", orderID)

	if err := c.orderInteractor.CancelOrder(orderID); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	log.Printf("Order %d cancelled successfully", orderID)
	return nil
}

// Start starts consuming messages from the queue
// This is a simplified example; in production, you'd connect to a real message queue
func (c *OrderConsumer) Start() error {
	log.Println("Starting order consumer...")

	// TODO: Connect to message queue (RabbitMQ, Kafka, etc.)
	// Example:
	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	// if err != nil {
	//     return err
	// }
	// defer conn.Close()
	//
	// ch, err := conn.Channel()
	// if err != nil {
	//     return err
	// }
	// defer ch.Close()
	//
	// msgs, err := ch.Consume(
	//     "order_queue", // queue name
	//     "",            // consumer
	//     true,          // auto-ack
	//     false,         // exclusive
	//     false,         // no-local
	//     false,         // no-wait
	//     nil,           // args
	// )
	// if err != nil {
	//     return err
	// }
	//
	// for msg := range msgs {
	//     if err := c.ConsumeMessage(msg.Body); err != nil {
	//         log.Printf("Error processing message: %v", err)
	//     }
	// }

	log.Println("Order consumer is running...")

	// Block forever
	select {}
}
