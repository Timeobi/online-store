package middleware

import (
	"net/http"

	"github.com/Timeobi/go-ecommerce/internal/model"
)

// RequireRole — middleware, которое пропускает запрос дальше только если
// роль пользователя (уже определённая через Auth middleware) совпадает с требуемой.
func RequireRole(role model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleContextKey).(model.Role)
			if !ok {
				respondUnauthorized(w, "не удалось определить роль пользователя")
				return
			}

			if userRole != role {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				if _, err := w.Write([]byte(`{"error": "недостаточно прав для выполнения операции"}`)); err != nil {
					_ = err
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
