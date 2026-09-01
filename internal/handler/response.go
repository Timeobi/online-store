package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

var log *slog.Logger

// SetLogger должен быть вызван один раз при старте приложения (см. main.go),
// чтобы хендлеры могли логировать ошибки через общий структурированный логгер.
func SetLogger(l *slog.Logger) {
	log = l
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(w http.ResponseWriter, status int, message string) {
	// Логируем только серверные ошибки (5xx) — клиентские ошибки (4xx, вроде
	// "невалидный JSON" или "не найдено") — это ожидаемая часть нормальной
	// работы API, не повод засорять логи уровня Error.
	if status >= 500 && log != nil {
		log.Error("request failed", slog.Int("status", status), slog.String("message", message))
	}
	respondJSON(w, status, errorResponse{Error: message})
}
