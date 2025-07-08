package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/db"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/handler"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/repository"
	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/service"
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

	repo := repository.NewProductRepository(dbpool)
	svc := service.NewProductService(repo)
	h := handler.NewProductHandler(svc)

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	router.HandleFunc("/products", h.Create).Methods("POST")
	router.HandleFunc("/products", h.GetAll).Methods("GET")
	router.HandleFunc("/products/{id}", h.GetByID).Methods("GET")
	router.HandleFunc("/products/{id}", h.Delete).Methods("DELETE")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Product service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
