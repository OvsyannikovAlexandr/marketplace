package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/db"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/handler"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env")
	}

	ctx := context.Background()

	db, err := db.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	router := mux.NewRouter()

	router.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
