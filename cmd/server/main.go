package main

import (
	"bytes"
	"codejudger/db"
	"codejudger/db/query"
	"codejudger/internal/hackacode"
	"codejudger/internal/judger"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "codejudger/cmd/server/docs"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	httpSwagger "github.com/swaggo/http-swagger"
)

type LanguageConfig struct {
	Compile   string   `json:"compile"`
	Extension string   `json:"extension"`
	Run       []string `json:"run"`
	File      string   `json:"file"`
}

var Languages = map[string]LanguageConfig{
	"C++": {
		Extension: "cpp",
		File:      "main.cpp",
		Compile:   "/usr/bin/g++ -O2 -o main main.cpp -Wall",
		Run:       []string{"./main"},
	},
	"C": {
		Extension: "c",
		File:      "main.c",
		Compile:   "/usr/bin/gcc -O2 -o main main.c -Wall",
		Run:       []string{"./main"},
	},
	"Rust": {
		Extension: "rs",
		File:      "main.rs",
		Compile:   "rustc main.rs -o main",
		Run:       []string{"./main"},
	},
	"Go": {
		Extension: "go",
		File:      "main.go",
		Compile:   "go build -o main main.go",
		Run:       []string{"./main"},
	},
	"Python": {
		Extension: "py",
		File:      "main.py",
		Run:       []string{"/usr/bin/python3", "main.py"},
		Compile:   "",
	},
	"Javascript": {
		Extension: "js",
		File:      "main.js",
		Run:       []string{"/usr/bin/node", "main.js"},
	},
	"Ruby": {
		Extension: "rb",
		File:      "main.rb",
		Run:       []string{"ruby", "main.rb"},
	},
	"PHP": {
		Extension: "php",
		File:      "main.php",
		Run:       []string{"php", "main.php"},
	},
	"C#": {
		Extension: "cs",
		File:      "main.cs",
		Compile:   "dotnet build -o out main.cs",
		Run:       []string{"dotnet", "out/main.dll"},
	},
}

type RequestData struct {
	Code     string `json:"code"`
	Slug     string `json:"slug"`
	Language string `json:"language"`
}

func main() {
	http.HandleFunc("/get-token", hackacode.ApiHandler)
	fmt.Println("hello! this is hackacode/s code judger")
	http.HandleFunc("/api/v1", apiHandler)
	http.HandleFunc("/api/v1/run", hackacode.RunHandler)
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	port := "0.0.0.0:1072"

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// @Summary      Judge code
// @Description  Receives code, language, and problem slug, runs the code, and returns the judge results.
// @Tags         judge
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        request body RequestData true "Code and problem data"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      401 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Success      200 {object} judger.JudgeResponse
// @Router       /api/v1 [post]
func apiHandler(w http.ResponseWriter, r *http.Request) {

	// i freaking hate cors
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

	langCfg, exists := Languages[language]
	if !exists {
		http.Error(w, "unsupported language", http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "code cannot be empty", http.StatusBadRequest)
		return
	}

	var judgerTestCases []judger.TestCase
	for _, tc := range test_cases {
		tcBytes, _ := json.Marshal(tc)
		var testCase judger.TestCase
		if err := json.Unmarshal(tcBytes, &testCase); err == nil {
			judgerTestCases = append(judgerTestCases, testCase)
		}
	}

	judgerConfig := judger.IsolateConfig{
		File:        langCfg.File,
		Code:        requestData.Code,
		Run:         langCfg.Run,
		Compile:     langCfg.Compile,
		TestCases:   judgerTestCases,
		Token:       authHeader[7:],
		MemoryLimit: int(challenge["memory_limit"].(float64)),
		TimeLimit:   int(challenge["time_limit"].(float64)),
	}

	results, err := judger.RunIsolate(judgerConfig)
	fmt.Println("results:", results)
	fmt.Println("error:", err)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"status":  "comp-failed",
			"message": fmt.Sprintf("%v", err),
			"id":      uuid.New().String(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	if len(results) == 0 {
		http.Error(w, "no judge results returned", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	status := "ACCEPTED"

	for _, result := range results {
		if !result.Passed {
			status = "FAILED"
			break
		}
	}

	passedCount := 0
	for _, result := range results {
		if result.Passed {
			passedCount++
		}
	}
	passedPercentage := float64(passedCount) / float64(len(results)) * 100

	resp := map[string]interface{}{
		"slug":     requestData.Slug,
		"language": requestData.Language,
		"code":     requestData.Code,
		"status":   status,
		"results":  results,
		"score":    passedPercentage,
		"id":       uuid.New().String(),
	}

	user, _ := query.GetUserByJWT(authHeader[7:])

	if user != nil {
		var submissions []map[string]interface{}
		if user.Submissions != "" {
			_ = json.Unmarshal([]byte(user.Submissions), &submissions)
		}

		newSubmission := map[string]interface{}{
			"challenge": challenge["slug"],
			"code":      requestData.Code,
			"result":    resp,
			"language":  requestData.Language,
			"timestamp": time.Now().Format(time.RFC3339),
			"status":    resp["status"],
			"score":     resp["score"],
			"duelId":    nil,
			"id":        uuid.New().String(),
		}

		submissions = append(submissions, newSubmission)

		submissionsJSON, _ := json.Marshal(submissions)
		user.Submissions = string(submissionsJSON)

		client := db.CreateClient()
		_, _, err := client.From("users").
			Update(map[string]interface{}{"submissions": json.RawMessage(submissionsJSON)}, "id", "eq").
			Eq("id", user.ID).
			Execute()
		if err != nil {
			fmt.Println("error updating user submissions:", err)
		}

		sendSlackNotification(user, challenge, newSubmission)
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

func sendSlackNotification(user *query.User, challenge map[string]interface{}, submission map[string]interface{}) {
	webhookURL := db.GetEnvVar("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	status := submission["status"].(string)
	score := submission["score"].(float64)

	statusEmoji := "❌"
	if status == "ACCEPTED" {
		statusEmoji = "✅"
	}

	message := map[string]interface{}{
		"text": fmt.Sprintf("%s %s: %s submitted by %s - %.0f%%", statusEmoji, challenge["slug"], status, user.Email, score),
	}

	payload, _ := json.Marshal(message)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return
	}
	defer resp.Body.Close()
}
