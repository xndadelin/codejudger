package hackacode

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/joho/godotenv"
)

type Response struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		http.Error(w, "Failed to load environment variables", http.StatusInternalServerError)
		return
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		http.Error(w, "JWT_SECRET is not set", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		APIKey string `json:"api_key"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestBody)

	if err != nil || requestBody.APIKey == "" {
		http.Error(w, "No api_key provided", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(10 * 365 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"sub": requestBody.APIKey,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := Response{
		Token:   tokenString,
		Message: "Token generated successfully",
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to create JSON response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}
