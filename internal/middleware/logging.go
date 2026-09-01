package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger — middleware, которое логирует каждый HTTP-запрос структурированно
// через slog, включая уникальный request_id для сквозной трассировки.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// chi/middleware.RequestID должен быть подключён РАНЬШЕ этого middleware
			// в цепочке (см. main.go) — он генерирует и кладёт ID в контекст.
			requestID := middleware.GetReqID(r.Context())

			// Оборачиваем ResponseWriter, чтобы после обработки запроса
			// узнать, какой статус-код реально был отправлен клиенту —
			// стандартный http.ResponseWriter сам по себе этого не раскрывает.
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Кладём request_id в заголовок ответа — клиент (или служба поддержки)
			// сможет сослаться на этот ID при обращении с проблемой.
			w.Header().Set("X-Request-ID", requestID)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			log.Info("request completed",
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", duration),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}
