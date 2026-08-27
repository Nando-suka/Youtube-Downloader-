package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Simple in-memory metrics suitable untuk basic observability.
type routeMetrics struct {
	Count        int64   `json:"count"`
	AvgDuration  float64 `json:"avg_duration_ms"`
	LastDuration int64   `json:"last_duration_ms"`
}

var (
	metricsMu sync.Mutex
	// key: method + " " + path
	metricsByRoute = make(map[string]*routeMetrics)
)

// MetricsMiddleware mengumpulkan metrik dasar untuk setiap request.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		key := method + " " + path

		metricsMu.Lock()
		m, ok := metricsByRoute[key]
		if !ok {
			m = &routeMetrics{}
			metricsByRoute[key] = m
		}
		m.Count++
		m.LastDuration = duration.Milliseconds()
		// simple running average
		m.AvgDuration = m.AvgDuration + (float64(m.LastDuration)-m.AvgDuration)/float64(m.Count)
		metricsMu.Unlock()
	})
}

// MetricsHandler meng-expose snapshot metrik dalam bentuk JSON.
func MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metricsMu.Lock()
		snapshot := make(map[string]routeMetrics, len(metricsByRoute))
		for k, v := range metricsByRoute {
			snapshot[k] = *v
		}
		metricsMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(snapshot)
	})
}

