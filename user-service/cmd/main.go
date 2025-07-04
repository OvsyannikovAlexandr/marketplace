package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/db"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/handler"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/repository"
	"github.com/OvsyannikovAlexandr/marketplace/user-service/internal/service"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Ошибка загрузки .env")
	}

	ctx := context.Background()

	dbpool, err := db.NewDatabase(ctx)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer dbpool.Close()
	log.Println("Connected to PostgreSQL")

	userRepo := repository.NewUserRepository(dbpool)
	authService := service.NewAuthService(userRepo)
	authHendler := handler.NewAuthHandler(authService)

	router := mux.NewRouter()
	router.HandleFunc("/register", authHendler.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", authHendler.LoginHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Println("Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
