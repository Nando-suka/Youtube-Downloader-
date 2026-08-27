package handlers

import (
	"encoding/json"
	"net/http"

	"Youtube_donwloader/config"
)

// HealthHandler memberikan informasi sederhana tentang status aplikasi.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()

	status := struct {
		Status           string `json:"status"`
		HasYoutubeAPIKey bool   `json:"has_youtube_api_key"`
	}{
		Status:           "ok",
		HasYoutubeAPIKey: len(cfg.YoutubeAPIKeys) > 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}












