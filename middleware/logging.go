package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// statusRecorder membungkus ResponseWriter untuk menangkap status dan ukuran respons.
type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// LoggingMiddleware menambahkan logging terstruktur untuk setiap request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-Id", reqID)

		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		clientIP := getClientIP(r)

		entry := map[string]interface{}{
			"ts":         start.Format(time.RFC3339Nano),
			"req_id":     reqID,
			"ip":         clientIP,
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     rec.status,
			"size":       rec.size,
			"latency_ms": duration.Milliseconds(),
			"user_agent": r.UserAgent(),
		}

		data, err := json.Marshal(entry)
		if err != nil {
			log.Printf("request_log_error: %v", err)
			return
		}

		log.Println(string(data))
	})
}
