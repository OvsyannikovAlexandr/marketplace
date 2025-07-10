package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/cache"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/db"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/handler"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/repository"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/internal/service"
	"github.com/OvsyannikovAlexandr/marketplace/order-service/pkg/kafka"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file")
	}

	ctx := context.Background()

	dbpool, err := db.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbpool.Close()

	fmt.Println("Connected to PostgreSQl")

	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := "logs"

	producer := kafka.NewOrderProducer(kafkaBroker, kafkaTopic)
	redisAddr := os.Getenv("REDIS_ADDR")
	redisCache := cache.NewRedisCache(redisAddr)

	orderRepo := repository.NewOrderRepository(dbpool)
	orderService := service.NewOrderService(orderRepo, producer, redisCache)
	orederHandler := handler.NewOrderHandler(orderService)

	router := mux.NewRouter()

	router.HandleFunc("/orders", orederHandler.Create).Methods("POST")
	router.HandleFunc("/orders", orederHandler.GetAll).Methods("GET")
	router.HandleFunc("/orders/{id}", orederHandler.GetByID).Methods("GET")
	router.HandleFunc("/orders/{id}", orederHandler.Delete).Methods("DELETE")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Order service OK"))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
