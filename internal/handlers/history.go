package handlers

import (
	"net/http"
	"sort"
	"sync"
	"time"

	"Youtube_donwloader/internal/jobs"
)

var (
	downloadHistory []*HistoryEntry
	historyMutex    sync.RWMutex
	maxHistorySize  = 50 // Simpan maksimal 50 riwayat terakhir
)

// HistoryEntry merepresentasikan entry dalam riwayat download
type HistoryEntry struct {
	JobID     string    `json:"job_id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	FileURL   string    `json:"file_url,omitempty"`
}

// AddToHistory menambahkan entry ke riwayat
func AddToHistory(job *jobs.DownloadJob) {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	entry := &HistoryEntry{
		JobID:     job.ID,
		URL:       job.URL,
		Status:    string(job.Status),
		CreatedAt: job.CreatedAt,
		FinishedAt: job.FinishedAt,
	}

	if job.Status == jobs.StatusCompleted {
		entry.FileURL = job.FileURL
	}

	// Cek apakah sudah ada, update jika ada
	found := false
	for i, e := range downloadHistory {
		if e.JobID == job.ID {
			downloadHistory[i] = entry
			found = true
			break
		}
	}

	if !found {
		downloadHistory = append(downloadHistory, entry)
	}

	// Sort by CreatedAt descending (terbaru di atas)
	sort.Slice(downloadHistory, func(i, j int) bool {
		return downloadHistory[i].CreatedAt.After(downloadHistory[j].CreatedAt)
	})

	// Batasi ukuran history
	if len(downloadHistory) > maxHistorySize {
		downloadHistory = downloadHistory[:maxHistorySize]
	}
}

// GetHistory mengambil riwayat download
func GetHistory() []*HistoryEntry {
	historyMutex.RLock()
	defer historyMutex.RUnlock()

	// Return copy untuk thread safety
	result := make([]*HistoryEntry, len(downloadHistory))
	copy(result, downloadHistory)
	return result
}

// HistoryHandler menangani request riwayat download
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	history := GetHistory()
	SendSuccess(w, map[string]interface{}{
		"history": history,
		"count":   len(history),
	}, http.StatusOK)
}
