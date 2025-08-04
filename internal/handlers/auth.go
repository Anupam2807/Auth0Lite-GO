package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Anupam2807/go-auth-service/internal/db"
	"github.com/Anupam2807/go-auth-service/internal/types"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func Welcome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Welcome to Go Auth Lite"}
	json.NewEncoder(w).Encode(response)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = validate.Struct(user)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`
	err = db.DB.QueryRow(checkQuery, user.Email).Scan(&exists)

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id`
	err = db.DB.QueryRow(query, user.Email, user.Password, user.Role).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Error Saving User", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User Created Successfully",
		"userId":  user.ID,
	})
}
