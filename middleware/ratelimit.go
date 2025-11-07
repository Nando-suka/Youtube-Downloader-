package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.Allow(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Clean up old requests
	if requests, exists := rl.requests[ip]; exists {
		var validRequests []time.Time
		for _, t := range requests {
			if now.Sub(t) <= rl.window {
				validRequests = append(validRequests, t)
			}
		}
		rl.requests[ip] = validRequests
	}

	// Check if under limit
	if len(rl.requests[ip]) >= rl.limit {
		return false
	}

	// Record new request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func getClientIP(r *http.Request) string {
	// Check for forwarded IP first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the list (client IP)
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	// Fallback to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
