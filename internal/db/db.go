package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DB_URL"))

	if err != nil {
		log.Fatal("DB connection failed:", err)

	}
	err = DB.Ping()
	if err != nil {
		log.Fatal("DB ping failed:", err)
	}
	log.Println("Connected to PostgreSQL DB")

}
