package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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
			Thumbnail6s struct {
				Default struct {
					URL string `json:"url"`
				} `json:"default"`
			} `json:"thumbnails"`
		} `json:"snippet"`
	} `json:"items"`
}

var (
	apiKeyManager *config.APIKeyManager
	kmOnce        sync.Once
)

func getAPIKeyManager() *config.APIKeyManager {
	kmOnce.Do(func() {
		cfg := config.Load()
		apiKeyManager = config.NewAPIKeyManager(cfg.YoutubeAPIKeys)
	})
	return apiKeyManager
}

// isSafeQuery memvalidasi query string untuk keamanan
func isSafeQuery(q string) bool {
	// Jika query kosong, tidak aman
	if q == "" {
		return false
	}

	// Trim spasi di awal dan akhir
	q = strings.TrimSpace(q)

	// Cek panjang query
	if len(q) > 200 {
		return false
	}

	// Daftar karakter yang diizinkan
	allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	allowedChars += "-_.,'\":!?@#$%^&*()[]{}<>|\\/+=~` "
	allowedChars += "áéíóúÁÉÍÓÚäëïöüÄËÏÖÜñÑ" // Karakter internasional

	// Validasi setiap karakter
	for _, char := range q {
		if !strings.ContainsRune(allowedChars, char) {
			return false
		}
	}

	// Cek apakah query hanya berisi karakter whitespace
	if strings.TrimSpace(q) == "" {
		return false
	}

	// Cek kata-kata berbahaya (opsional)
	badWords := []string{"<script>", "javascript:", "onload", "onerror", "alert("}
	for _, word := range badWords {
		if strings.Contains(strings.ToLower(q), strings.ToLower(word)) {
			return false
		}
	}

	return true
}

// SearchHandler menangani request pencarian video YouTube
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Validasi method
	if r.Method != http.MethodGet {
		SendError(w, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// Ambil query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		SendError(w, "MISSING_QUERY", "Parameter 'q' wajib diisi", http.StatusBadRequest)
		return
	}

	// Validasi query
	if !isSafeQuery(query) {
		SendError(w, "INVALID_QUERY", "Query mengandung karakter tidak valid atau terlalu panjang", http.StatusBadRequest)
		return
	}

	// Ambil API key manager
	manager := getAPIKeyManager()
	if manager == nil || manager.GetAvailableKeysCount() == 0 {
		SendError(w, "API_KEYS_NOT_CONFIGURED", "YouTube API keys belum dikonfigurasi", http.StatusInternalServerError)
		return
	}

	// Setup context dengan timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	attempts := manager.GetAvailableKeysCount()
	var lastErr error
	var lastStatus int
	var result *SearchResult

	// Coba dengan semua API key yang tersedia
	for i := 0; i < attempts; i++ {
		apiKey := manager.GetCurrentKey()
		if apiKey == "" {
			lastErr = fmt.Errorf("no API key available")
			manager.RotateKey()
			continue
		}

		// Bangun URL API
		apiURL := fmt.Sprintf(
			"https://www.googleapis.com/youtube/v3/search?part=snippet&maxResults=10&q=%s&type=video&key=%s",
			url.QueryEscape(query),
			apiKey,
		)

		// Buat request dengan context
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
		if err != nil {
			lastErr = err
			break
		}

		// Tambahkan headers
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Accept", "application/json")

		// Kirim request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			break
		}

		// Baca response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			break
		}

		lastStatus = resp.StatusCode

		// Handle API errors
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			// Quota habis atau rate limit, rotate key dan coba lagi
			fmt.Printf("API Key %d quota exceeded (status %d), rotating...\n",
				manager.GetCurrentIndex(), resp.StatusCode)
			manager.RotateKey()
			continue
		}

		// Cek error lain
		if resp.StatusCode >= 400 {
			// Coba parse error message
			var apiError struct {
				Error struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
					Errors  []struct {
						Domain  string `json:"domain"`
						Reason  string `json:"reason"`
						Message string `json:"message"`
					} `json:"errors"`
				} `json:"error"`
			}

			if err := json.Unmarshal(body, &apiError); err == nil {
				lastErr = fmt.Errorf("YouTube API error: %s", apiError.Error.Message)
			} else {
				lastErr = fmt.Errorf("YouTube API returned status %d", resp.StatusCode)
			}

			// Jika bukan quota error, langsung return
			if !strings.Contains(strings.ToLower(string(body)), "quota") {
				break
			}

			// Jika quota error, rotate key dan coba lagi
			manager.RotateKey()
			continue
		}

		// Parse successful response
		var searchResult SearchResult
		if err := json.Unmarshal(body, &searchResult); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %v", err)
			break
		}

		result = &searchResult
		manager.RecordUsage() // Catat penggunaan yang sukses
		break
	}

	// Handle hasil
	if result != nil {
		w.Header().Set("X-API-Key-Index", fmt.Sprintf("%d", manager.GetCurrentIndex()))
		w.Header().Set("X-API-Key-Usage", fmt.Sprintf("%v", manager.GetUsageStats()))
		SendSuccess(w, result, http.StatusOK)
		return
	}

	// Jika sampai sini, berarti semua percobaan gagal
	if lastErr != nil {
		SendError(w, "SEARCH_FAILED", fmt.Sprintf("Gagal mencari video: %v", lastErr), http.StatusInternalServerError)
		return
	}

	SendError(w, "SEARCH_FAILED", fmt.Sprintf("Gagal mencari video setelah %d percobaan (status terakhir: %d)", attempts, lastStatus), http.StatusInternalServerError)
}
