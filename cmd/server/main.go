package main

import (
	"codejudger/db"
	"codejudger/internal/judger"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type RequestData struct {
	Code     string `json:"code"`
	Slug     string `json:"slug"`
	Language string `json:"language"`
}

func main() {
	fmt.Println("hello! this is hackacode/s code judger")
	http.HandleFunc("/api/v1", apiHandler)
	http.ListenAndServe("0.0.0.0:3000", nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !isAuthorized(authHeader) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestData RequestData
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if requestData.Code == "" || requestData.Slug == "" || requestData.Language == "" {
		http.Error(w, "oopssie!! you forgot to provide some data", http.StatusBadRequest)
		return
	}

	client := db.CreateClient()
	data, _, err := client.From("problems").Select("*", "", false).
		Eq("slug", requestData.Slug).
		Execute()
	if err != nil {
		http.Error(w, "there has been an error in fetching the challenge! please try again later or contact support", http.StatusInternalServerError)
		return
	}

	var challenges []map[string]interface{}
	if err := json.Unmarshal(data, &challenges); err != nil {
		http.Error(w, "i cant parse the challenge data", http.StatusInternalServerError)
		return
	}
	if len(challenges) == 0 {
		http.Error(w, "challenge not found", http.StatusNotFound)
		return
	}
	challenge := challenges[0]

	judgerConfig := judger.IsolateConfig{
		BoxID:   1,
		Memory:  256,
		Runtime: 5,
		Command: fmt.Sprintf("echo '%s' | %s", requestData.Code, requestData.Language),
	}
	if err := judger.RunIsolate(judgerConfig); err != nil {
		http.Error(w, fmt.Sprintf("oh no!!! judger error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"title":       challenge["title"],
		"description": challenge["description"],
		"language":    requestData.Language,
		"code":        requestData.Code,
		"status":      "success",
		"message":     "your code has been successfully judged",
	}
	json.NewEncoder(w).Encode(resp)
}

func isAuthorized(authHeader string) bool {
	return authHeader != "" && len(authHeader) >= 7 && authHeader[:7] == "Bearer " && verifyToken(authHeader[7:])
}

func verifyToken(tokenString string) bool {
	secretKey := []byte(db.GetEnvVar("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	return err == nil && token.Valid
}
