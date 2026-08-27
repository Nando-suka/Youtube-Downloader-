package main

import (
	"Youtube_donwloader/config"
	"Youtube_donwloader/internal/handlers"
	"Youtube_donwloader/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	// Initialize download queue
	handlers.InitDownloadQueue()

	r := mux.NewRouter()

	// Static files dengan no directory listing (harus sebelum middleware)
	staticHandler := http.StripPrefix("/staticDua/",
		middleware.NoDirectoryListingHandler(
			http.Dir("staticDua"),
			"",
		),
	)
	r.PathPrefix("/staticDua/").Handler(staticHandler)

	// Routes
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/download", handlers.DownloadHandler).Methods("POST")
	r.HandleFunc("/download/{job_id}/status", handlers.DownloadStatusHandler).Methods("GET")
	r.HandleFunc("/download/{job_id}/file", handlers.DownloadFileHandler).Methods("GET")
	r.HandleFunc("/search", handlers.SearchHandler).Methods("GET")
	r.HandleFunc("/history", handlers.HistoryHandler).Methods("GET")
	r.HandleFunc("/healthz", handlers.HealthHandler).Methods("GET")
	r.Handle("/metrics", middleware.MetricsHandler()).Methods("GET")

	// Apply middlewares (order matters: security -> metrics -> logging -> rate limit -> handlers)
	// Static files tetap dapat diakses karena sudah di-register sebelum middleware
	r.Use(middleware.SecurityHeadersMiddleware)
	r.Use(middleware.MetricsMiddleware)
	r.Use(middleware.LoggingMiddleware)

	// Apply rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	r.Use(rateLimiter.Middleware)

	// Tempat untuk memulai (Starting point)
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
