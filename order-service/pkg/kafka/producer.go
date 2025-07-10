package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type OrderProducer struct {
	writer *kafka.Writer
}

type Producer interface {
	SendOrderCreated(ctx context.Context, event OrderCreatedEvent) error
}

type OrderCreatedEvent struct {
	OrderID    int64   `json:"order_id"`
	UserID     int64   `json:"user_id"`
	ProductIDs []int64 `json:"product_ids"`
	TotalPrice float64 `json:"total_price"`
	CreatedAt  string  `json:"created_at"`
}

func NewOrderProducer(brokerAddress, topic string) *OrderProducer {
	return &OrderProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokerAddress),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *OrderProducer) SendOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("order_id"),
		Value: msg,
		Time:  time.Now(),
	})
}
