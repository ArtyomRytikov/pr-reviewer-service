package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/storage"
	"time"

	_ "github.com/lib/pq"
)

type Server struct {
	service *service.PRService
}

func NewServer(service *service.PRService) *Server {
	return &Server{service: service}
}

func (s *Server) handleTeamAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.service.CreateTeam(&team); err != nil {
		if err.Error() == "TEAM_EXISTS" {
			sendErrorResponse(w, "TEAM_EXISTS", "team_name already exists", http.StatusBadRequest)
			return
		}
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"team": team})
}

func (s *Server) handleTeamGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		sendError(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := s.service.GetTeam(teamName)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			sendErrorResponse(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
			return
		}
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (s *Server) handleSetUserActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.service.SetUserActive(req.UserID, req.IsActive)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			sendErrorResponse(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
			return
		}
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": user})
}

func (s *Server) handleCreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := s.service.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "PR_EXISTS" {
			sendErrorResponse(w, "PR_EXISTS", "PR id already exists", http.StatusConflict)
		} else if errMsg == "NOT_FOUND" {
			sendErrorResponse(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
		} else {
			sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"pr": pr})
}

func (s *Server) handleMergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, err := s.service.MergePR(req.PullRequestID)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			sendErrorResponse(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
			return
		}
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pr": pr})
}

func (s *Server) handleReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pr, newUserID, err := s.service.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "NOT_FOUND" {
			sendErrorResponse(w, "NOT_FOUND", "resource not found", http.StatusNotFound)
		} else if errMsg == "PR_MERGED" {
			sendErrorResponse(w, "PR_MERGED", "cannot reassign on merged PR", http.StatusConflict)
		} else if errMsg == "NOT_ASSIGNED" {
			sendErrorResponse(w, "NOT_ASSIGNED", "reviewer is not assigned to this PR", http.StatusConflict)
		} else if errMsg == "NO_CANDIDATE" {
			sendErrorResponse(w, "NO_CANDIDATE", "no active replacement candidate in team", http.StatusConflict)
		} else {
			sendError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pr":          pr,
		"replaced_by": newUserID,
	})
}

func (s *Server) handleGetUserReviewPRs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		sendError(w, "user_id is required", http.StatusBadRequest)
		return
	}

	prs, err := s.service.GetUserReviewPRs(userID)
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.service.GetStats()
	if err != nil {
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", s.handleTeamAdd)
	mux.HandleFunc("/team/get", s.handleTeamGet)
	mux.HandleFunc("/users/setIsActive", s.handleSetUserActive)
	mux.HandleFunc("/users/getReview", s.handleGetUserReviewPRs)
	mux.HandleFunc("/pullRequest/create", s.handleCreatePR)
	mux.HandleFunc("/pullRequest/merge", s.handleMergePR)
	mux.HandleFunc("/pullRequest/reassign", s.handleReassignReviewer)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/health", s.handleHealth)

	return mux
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func sendErrorResponse(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := models.ErrorResponse{}
	errorResp.Error.Code = code
	errorResp.Error.Message = message

	json.NewEncoder(w).Encode(errorResp)
}

func main() {
	var dbStorage *storage.PostgresStorage
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "postgres"),
			getEnv("DB_PASSWORD", "password"),
			getEnv("DB_NAME", "pr_reviewer"),
		)

		dbStorage, err = storage.NewPostgresStorage(connStr)
		if err == nil {
			break
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database after retries:", err)
	}
	defer dbStorage.Close()

	prService := service.NewPRService(dbStorage)

	server := NewServer(prService)

	router := server.SetupRoutes()

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
