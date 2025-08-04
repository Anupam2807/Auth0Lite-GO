package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Anupam2807/go-auth-service/internal/db"
	"github.com/Anupam2807/go-auth-service/internal/types"
	"github.com/Anupam2807/go-auth-service/internal/utils"
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
	user.Provider = `email`
	query := `INSERT INTO users (email, password, role,provider) VALUES ($1, $2, $3,$4) RETURNING id`
	err = db.DB.QueryRow(query, user.Email, user.Password, user.Role, user.Provider).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Error Saving User", http.StatusBadRequest)
		log.Println("error:", err)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User Created Successfully",
		"userId":  user.ID,
	})
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, email, provider, role FROM users")
	if err != nil {
		http.Error(w, "Failed to query users", http.StatusInternalServerError)
		log.Println("DB error:", err)
		return
	}
	defer rows.Close()

	var users []types.User

	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.ID, &user.Email, &user.Provider, &user.Role)
		if err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			log.Println("Scan error:", err)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading rows", http.StatusInternalServerError)
		log.Println("Row error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"users": users,
		"count": len(users),
	}
	json.NewEncoder(w).Encode(response)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {

	var input types.LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		log.Println(err)
		return
	}
	err = validate.Struct(input)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		http.Error(w, validationErrors.Error(), http.StatusBadRequest)
	}

	var dbUser types.User
	query := `SELECT id, email, password, role, provider FROM users WHERE email=$1`
	err = db.DB.QueryRow(query, input.Email).Scan(
		&dbUser.ID,
		&dbUser.Email,
		&dbUser.Password,
		&dbUser.Role,
		&dbUser.Provider,
	)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(input.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	accessToken, refreshToken, err := utils.GenerateTokens(int(dbUser.ID), dbUser.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": map[string]interface{}{
			"id":       dbUser.ID,
			"email":    dbUser.Email,
			"role":     dbUser.Role,
			"provider": dbUser.Provider,
		},
	})
}
