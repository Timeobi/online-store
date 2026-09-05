package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter хранит отдельный "лимитер" для каждого IP-адреса.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

// NewRateLimiter создаёт ограничитель: r — сколько запросов в секунду разрешено
// в среднем, burst — сколько запросов можно сделать "залпом" перед тем, как
// ограничение начнёт действовать.
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		r:        r,
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// Limit — middleware, ограничивающее частоту запросов по IP-адресу клиента.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // на случай, если формат адреса неожиданный
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			if _, err := w.Write([]byte(`{"error": "слишком много попыток, повторите позже"}`)); err != nil {
				_ = err
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
