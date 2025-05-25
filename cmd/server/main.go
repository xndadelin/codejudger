package main

import (
	"codejudger/db"
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
	/* 	client := db.CreateClient()

	   	data, count, err := client.From("problems").Select("*", "", false).Execute()

	   	if err != nil {
	   		fmt.Println("Error:", err)
	   		return
	   	}

	   	fmt.Println("Data:", string(data))
	   	fmt.Println("Count:", count)
	*/

	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		if !verifyToken(tokenString) {
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
		fmt.Println(requestData.Slug)
		data, _, err := client.From("problems").Select("*", "", false).
			Eq("slug", requestData.Slug).
			Execute()

		if err != nil {
			http.Error(w, "there has been an error in fetching the challenge! please try again later or contact support", http.StatusInternalServerError)
			return
		}

		if data == nil {
			http.Error(w, "challenge not found", http.StatusNotFound)
			return
		}

		var challenges []map[string]interface{}
		if err := json.Unmarshal(data, &challenges); err != nil {
			http.Error(w, "Error decoding challenge data", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"challenge": challenges,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe("0.0.0.0:3000", nil)
}

func verifyToken(tokenString string) bool {
	secretKey := []byte(db.GetEnvVar("JWT_SECRET"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return false
	}

	if !token.Valid {
		return false
	}

	return true
}
