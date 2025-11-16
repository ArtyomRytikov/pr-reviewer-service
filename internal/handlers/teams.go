package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/models"
	"pr-reviewer-service/internal/service"
)

type TeamHandler struct {
	service *service.PRService
}

func NewTeamHandler(service *service.PRService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateTeam(&team); err != nil {
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

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		sendError(w, "team_name is required", http.StatusBadRequest)
		return
	}

	team, err := h.service.GetTeam(teamName)
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
