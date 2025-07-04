package main

import (
	"fmt"
	"log"

	"github.com/OvsyannikovAlexandr/marketplace/migration-service/internal/migrate"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using defaults")
	}

	if err := migrate.Run(); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}

	fmt.Println("Миграции применены успешно")
}
