package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/models"
)

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
