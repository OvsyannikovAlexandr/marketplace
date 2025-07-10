package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/cache"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/db"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/handler"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/repository"
	"github.com/OvsyannikovAlexandr/marketplace/cart-service/internal/service"
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

	fmt.Println("Connected to PostgreSQL")

	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")

	redisAddr := os.Getenv("REDIS_ADDR")
	redisCache := cache.NewRedisCache(redisAddr)

	cartRepository := repository.NewCartRepository(dbpool)
	cartService := service.NewCartService(cartRepository, productServiceURL, redisCache)
	cartHandler := handler.NewCartHandler(cartService)

	router := mux.NewRouter()

	router.HandleFunc("/cart", cartHandler.AddItem).Methods("POST")
	router.HandleFunc("/cart/{user_id}", cartHandler.GetCartDetailsHandler).Methods("GET")
	router.HandleFunc("/cart/{user_id}/clear", cartHandler.ClearCart).Methods("DELETE")
	router.HandleFunc("/cart/{user_id}/checkout", cartHandler.Checkout).Methods("POST")
	router.HandleFunc("/cart/{user_id}/{product_id:[0-9]+}", cartHandler.DeleteItem).Methods("DELETE")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Cart service OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Cart service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
