package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bucket struct {
	tokens     int
	lastRefill time.Time
}

type RateLimiter struct {
	limit      int           // max tokens (burst)
	window     time.Duration // refill window
	mutex      sync.Mutex
	// key: route + ip
	buckets map[string]*bucket
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*bucket),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		routeKey := r.Method + ":" + r.URL.Path

		allowed, retryAfter := rl.allow(routeKey, clientIP)
		if !allowed {
			if retryAfter > 0 {
				w.Header().Set("Retry-After", formatRetryAfter(retryAfter))
			}
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(route, ip string) (bool, time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	key := route + "|" + ip

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.limit, lastRefill: now}
		rl.buckets[key] = b
	}

	// Refill based on elapsed time; simple fixed window refill.
	elapsed := now.Sub(b.lastRefill)
	if elapsed >= rl.window {
		b.tokens = rl.limit
		b.lastRefill = now
	}

	if b.tokens <= 0 {
		// time remaining until window resets
		retry := rl.window - elapsed
		if retry < 0 {
			retry = 0
		}
		return false, retry
	}

	b.tokens--
	return true, 0
}

func formatRetryAfter(d time.Duration) string {
	// prefer seconds integer for header
	secs := int(d.Seconds())
	if secs < 1 {
		secs = 1
	}
	return strconv.Itoa(secs)
}

func getClientIP(r *http.Request) string {
	// Prefer real client headers used by proxies
	for _, header := range []string{"CF-Connecting-IP", "X-Real-IP", "X-Forwarded-For"} {
		if value := r.Header.Get(header); value != "" {
			// X-Forwarded-For may contain list
			parts := strings.Split(value, ",")
			ip := strings.TrimSpace(parts[0])
			return normalizeIP(ip)
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return normalizeIP(r.RemoteAddr)
	}
	return normalizeIP(host)
}

func normalizeIP(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ip
	}
	return parsed.String()
}
