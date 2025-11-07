package main

import (
	"log"
	"net/http"

	"Youtube_donwloader/config"
	"Youtube_donwloader/internal/handlers"
	"Youtube_donwloader/middleware"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/staticDua/").Handler(http.StripPrefix("/staticDua/", http.FileServer(http.Dir("staticDua"))))

	// Routes
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/download", handlers.DownloadHandler).Methods("POST")
	r.HandleFunc("/search", handlers.SearchHandler).Methods("GET")

	// Apply rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	r.Use(rateLimiter.Middleware)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(http.ListenAndServe(":"+cfg.ServerPort, r))
}
