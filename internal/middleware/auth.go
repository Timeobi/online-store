package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Timeobi/go-ecommerce/internal/service"
)

// contextKey — отдельный тип для ключей контекста, чтобы избежать коллизий
// с ключами других пакетов (стандартная рекомендация Go для работы с context.Value).
type contextKey string

const (
	UserIDContextKey   contextKey = "user_id"
	UserRoleContextKey contextKey = "user_role"
)

// Auth — middleware, которое проверяет JWT-токен из заголовка Authorization.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondUnauthorized(w, "отсутствует заголовок Authorization")
				return
			}

			// Ожидаем формат "Bearer <токен>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondUnauthorized(w, "некорректный формат заголовка Authorization")
				return
			}
			tokenString := parts[1]

			claims, err := service.ParseToken(tokenString, jwtSecret)
			if err != nil {
				respondUnauthorized(w, "невалидный или истёкший токен")
				return
			}

			// Кладём данные пользователя в контекст запроса, чтобы хендлеры
			// дальше по цепочке могли получить, кто сейчас делает запрос.
			ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleContextKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte(`{"error": "` + message + `"}`)); err != nil {
		_ = err
	}
}
