package middleware

import (
	"log"
	"net/http"
	"rate-limiter/internal/rateLimiter"
)

// RateLimitMiddleware creates HTTP middleware for rate limiting
func RateLimitMiddleware(rl *rateLimiter.RateLimiter, keyExtractor func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyExtractor(r)

			if !rl.Allow(key) {
				// "If the rate limiter does NOT allow this key, then block the request"
				log.Printf("Rate limit exceeded - key: %s, path: %s, method: %s", key, r.URL.Path, r.Method)
				w.Header().Set("X-RateLimit-Exceeded", "true")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			log.Printf("Request allowed - key: %s, path: %s, method: %s", key, r.URL.Path, r.Method)
			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyExtractor extracts client IP as the rate limit key
func IPKeyExtractor(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

// APIKeyExtractor extracts API key from header
func APIKeyExtractor(r *http.Request) string {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		return IPKeyExtractor(r)
	}
	return key
}
