package main

import (
	"context"
	"log"

	"github.com/OvsyannikovAlexandr/marketplace/logging-service/internal/kafka"
)

func main() {
	log.Println("Starting Logging Service...")
	ctx := context.Background()

	if err := kafka.StartConsumer(ctx); err != nil {
		log.Fatalf("failed to start Kafka consumer: %v", err)
	}
}
