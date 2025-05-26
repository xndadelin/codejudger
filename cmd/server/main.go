package main

import (
	"codejudger/db"
	"codejudger/internal/judger"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type LanguageConfig struct {
	Compile   string `json:"compile"`
	Extension string `json:"extension"`
	Run       string `json:"run"`
	File      string `json:"file"`
}

var Languages = map[string]LanguageConfig{
	"C++": {
		Extension: "cpp",
		File:      "main.cpp",
		Compile:   "/usr/bin/g++ -O2 -o main main.cpp -Wall",
		Run:       "./main < input.txt > output.txt",
	},
	"C": {
		Extension: "c",
		File:      "main.c",
		Run:       "gcc -O2 -o main main.c -Wall 2> error.txt && ./main < input.txt > output.txt",
	},
	"C#": {
		Extension: "cs",
		File:      "main.cs",
		Run:       "dotnet new console -o main && cp main.cs main/Program.cs && cd main && dotnet build -c Release 2> ../error.txt && cd .. && cp -r main/bin/Release/net8.0/ ./program && ./program/main < input.txt > output.txt",
	},
	"Java": {
		Extension: "java",
		File:      "Main.java",
		Run:       "javac Main.java 2> error.txt && echo 'Main-Class: Main' > MANIFEST.MF && jar cfm Main.jar MANIFEST.MF Main.class && chmod +x Main.jar && java -jar Main.jar < input.txt > output.txt",
	},
	"Python": {
		Extension: "py",
		File:      "main.py",
		Run:       "/usr/bin/python3 main.py",
	},
	"Javascript": {
		Extension: "js",
		File:      "main.js",
		Run:       "node main.js < input.txt > output.txt",
	},
	"Ruby": {
		Extension: "rb",
		File:      "main.rb",
		Run:       "ruby main.rb < input.txt > output.txt",
	},
	"Rust": {
		Extension: "rs",
		File:      "main.rs",
		Run:       "rustc main.rs 2> error.txt && ./main < input.txt > output.txt",
	},
	"Go": {
		Extension: "go",
		File:      "main.go",
		Run:       "go mod init main 2> /dev/null && go build main.go 2> error.txt && ./main < input.txt > output.txt",
	},
	"PHP": {
		Extension: "php",
		File:      "main.php",
		Run:       "php main.php < input.txt > output.txt",
	},
}

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

	test_cases, ok := challenge["test_cases"].([]interface{})
	if !ok || len(test_cases) == 0 {
		http.Error(w, "oops! no test cases found for this challenge", http.StatusNotFound)
		return
	}

	language := requestData.Language
	code := requestData.Code

	if _, exists := Languages[language]; !exists {
		http.Error(w, "unsupported language", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "code cannot be empty", http.StatusBadRequest)
		return
	}

	judgerConfig := judger.IsolateConfig{
		File:    Languages[language].File,
		Code:    requestData.Code,
		Command: Languages[language].Run,
		Compile: Languages[language].Compile,
	}

	stdout, stderr, exitCode, err := judger.RunIsolate(judgerConfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("error running code: %v", err), http.StatusInternalServerError)
		return
	}

	exitCodeInt, convErr := strconv.Atoi(exitCode)
	if convErr != nil {
		http.Error(w, "failed to parse exit code", http.StatusInternalServerError)
		return
	}

	if exitCodeInt != 0 {
		http.Error(w, fmt.Sprintf("error in code execution: %s", stderr), http.StatusInternalServerError)
		return
	}
	if stdout == "" {
		http.Error(w, "no output produced by the code", http.StatusInternalServerError)
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
