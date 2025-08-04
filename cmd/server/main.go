package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/Anupam2807/go-auth-service/internal/db"
	"github.com/Anupam2807/go-auth-service/internal/handlers"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("⚠️ No .env file found or failed to load")
	}

	fmt.Println("DB URL:", os.Getenv("DB_URL"))

	db.Connect()

	router := http.NewServeMux()

	router.HandleFunc("GET /api", handlers.Welcome)
	router.HandleFunc("POST /api/user/register", handlers.RegisterUser)
	router.HandleFunc("GET /api/users", handlers.GetUsers)
	router.HandleFunc("POST /api/user/login", handlers.LoginUser)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on port", port)
	http.ListenAndServe(":"+port, router)
}
