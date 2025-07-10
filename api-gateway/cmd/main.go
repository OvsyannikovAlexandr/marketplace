package main

import (
	"log"
	"net/http"
	"os"

	"github.com/OvsyannikovAlexandr/marketplace/api-service/internal/proxy"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := proxy.NewRouter()
	log.Printf("API gateway running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
