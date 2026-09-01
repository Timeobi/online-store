package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
)

// Recovery — middleware, перехватывающее панику в любом хендлере ниже по цепочке.
// Логирует полную информацию (для разработчика), но отдаёт клиенту только
// общее сообщение об ошибке — детали реализации не должны утекать наружу.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := middleware.GetReqID(r.Context())

					log.Error("panic recovered",
						slog.String("request_id", requestID),
						slog.Any("error", rec),
						slog.String("stack", string(debug.Stack())),
						slog.String("path", r.URL.Path),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "внутренняя ошибка сервера"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
