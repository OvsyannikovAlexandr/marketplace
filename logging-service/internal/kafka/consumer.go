package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

const (
	kafkaBrokerAddress = "kafka:9092"
	topic              = "logs"
	groupID            = "logging-service"
)

func StartConsumer(ctx context.Context) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBrokerAddress},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	defer r.Close()

	log.Printf("Listening for messages on topic: %s\n", topic)
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}
		// Простой лог
		log.Printf("[Kafka] Received at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}
}
