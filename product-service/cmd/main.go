package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/product-service/internal/db"
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

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Product service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
