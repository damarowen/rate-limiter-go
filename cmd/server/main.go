package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"rate-limiter/internal/middleware"
	"rate-limiter/internal/rateLimiter"
)

// Helper to check if user is premium
func isPremiumUser(apiKey string) bool {
	if strings.Contains(apiKey, "premium") {
		return true
	} else {
		return false
	}
}

func main() {
	// Create default rate limiter: 10 requests per minute (basic users), default strategy
	defaultStrategy := rateLimiter.NewFixedWindowStrategy(10, time.Minute)
	limiter := rateLimiter.NewRateLimiter(defaultStrategy)

	// Configure premium tier: 100 requests per minute, second strategy
	premiumStrategy := rateLimiter.NewFixedWindowStrategy(100, time.Minute)
	apiPremiumKey := "premium-api-key"
	limiter.SetPremiumClient(apiPremiumKey, premiumStrategy)

	mux := http.NewServeMux()

	// Admin routes (NO rate limiting)
	mux.HandleFunc("/admin/reset", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("api_key")
		limiter.Reset(apiKey)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Rate limit reset for: " + apiKey))
	})

	// API routes (WITH rate limiting)
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		isPremium := isPremiumUser(apiKey)

		response := map[string]string{
			"message": "Hello! You are within rate limits.",
			"time":    time.Now().Format(time.RFC3339),
		}

		// Add premium user badge
		if isPremium {
			response["user_tier"] = "premium"
			response["rate_limit"] = "100 requests/minute"
			w.Header().Set("X-User-Tier", "premium")
		} else {
			response["user_tier"] = "basic"
			response["rate_limit"] = "10 requests/minute"
			w.Header().Set("X-User-Tier", "basic")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Apply rate limiting ONLY to /api/* routes
	mux.Handle("/api/", middleware.RateLimitMiddleware(limiter, middleware.APIKeyExtractor)(apiMux))

	log.Println("Server starting on :8080")
	log.Println("Rate limits:")
	log.Println("  - Basic users: 10 requests/minute")
	log.Println("  - Premium users (premium-api-key): 100 requests/minute")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
