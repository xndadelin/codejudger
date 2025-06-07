package hackacode

import (
	"codejudger/internal/judger"
	"encoding/json"
	"net/http"
)

type RunRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Input    string `json:"input"`
}

type RunResponse struct {
	Result judger.JudgeResult `json:"result"`
	Error  string             `json:"error,omitempty"`
}

func RunHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RunResponse{Error: "invalid request body"})
		return
	}
	if req.Code == "" || req.Language == "" || req.Input == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RunResponse{Error: "code, language, and input are required"})
		return
	}
	result, err := judger.RunSingleTest(req.Code, req.Language, req.Input)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(RunResponse{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RunResponse{Result: result})
}
