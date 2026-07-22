package handler

import (
	"encoding/json"
	"net/http"
)

// respondJSON пишет успешный ответ в формате JSON.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// errorResponse — стандартная структура для ошибок в API.
type errorResponse struct {
	Error string `json:"error"`
}

// respondError пишет ответ с ошибкой в формате JSON.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, errorResponse{Error: message})
}
