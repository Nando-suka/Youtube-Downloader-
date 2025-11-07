package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"Youtube_donwloader/config"
)

type SearchResult struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Channel     string `json:"channelTitle"`
			Description string `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Menggunakan API Key dari config
	cfg := config.Load()
	if len(cfg.YoutubeAPIKeys) == 0 {
		http.Error(w, "YouTube API keys not configured", http.StatusInternalServerError)
		return
	}
	apiKey := cfg.YoutubeAPIKeys[0] //gunakan key pertama

	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=10&q=%s&type=video&key=%s",
		url.QueryEscape(query),
		apiKey,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to search videos", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var result SearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Failed to parse results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
