package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/Anupam2807/go-auth-service/internal/db"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("⚠️ No .env file found or failed to load")
	}

	fmt.Println("DB URL:", os.Getenv("DB_URL"))

	db.Connect()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on port", port)
	http.ListenAndServe(":"+port, nil)
}
