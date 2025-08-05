package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Anupam2807/go-auth-service/internal/config"
	"github.com/Anupam2807/go-auth-service/internal/db"
	"github.com/Anupam2807/go-auth-service/internal/types"
	"github.com/Anupam2807/go-auth-service/internal/utils"
	"golang.org/x/oauth2"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := config.GoogleOAuthConfig.AuthCodeURL("random-state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Code exchange failed", http.StatusInternalServerError)
		return
	}
	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var googleUser types.GoogleUser

	err = json.Unmarshal(body, &googleUser)
	if err != nil || !googleUser.VerifiedEmail {
		http.Error(w, "Invalid user data", http.StatusUnauthorized)
		return
	}

	var dbUser types.User
	err = db.DB.QueryRow("SELECT id, email, role FROM users WHERE email=$1", googleUser.Email).Scan(
		&dbUser.ID,
		&dbUser.Email,
		&dbUser.Role,
	)

	if err != nil {

		dbUser.Email = googleUser.Email
		dbUser.Role = "user"
		dbUser.Provider = "google"

		err = db.DB.QueryRow(
			"INSERT INTO users (email, role, provider) VALUES ($1, $2, $3) RETURNING id",
			dbUser.Email, dbUser.Role, dbUser.Provider,
		).Scan(&dbUser.ID)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

	}
	accessToken, refreshToken, err := utils.GenerateTokens(int(dbUser.ID), dbUser.Email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Google login successful",
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
