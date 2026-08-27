package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"Youtube_donwloader/config"
	"Youtube_donwloader/internal/jobs"
)

var downloadQueue *jobs.Queue

// InitDownloadQueue menginisialisasi download queue (dipanggil dari main)
func InitDownloadQueue() {
	downloadQueue = jobs.NewQueue(3, jobs.ProcessDownload)
	// Set callback untuk update history ketika job selesai
	downloadQueue.SetOnJobDone(func(job *jobs.DownloadJob) {
		AddToHistory(job)
	})
}

// DownloadHandler menangani request download baru (async)
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	// Coba parse JSON dulu, jika gagal coba form
	if r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			SendError(w, "INVALID_REQUEST", "Request body tidak valid", http.StatusBadRequest)
			return
		}
	} else {
		req.URL = strings.TrimSpace(r.FormValue("url"))
	}

	if req.URL == "" {
		SendError(w, "MISSING_URL", "URL tidak boleh kosong", http.StatusBadRequest)
		return
	}

	if len(req.URL) > 2048 {
		SendError(w, "URL_TOO_LONG", "URL terlalu panjang (maksimal 2048 karakter)", http.StatusBadRequest)
		return
	}

	if !isValidURL(req.URL) {
		SendError(w, "INVALID_URL", "URL tidak valid. Hanya mendukung YouTube dan SoundCloud", http.StatusBadRequest)
		return
	}

	if downloadQueue == nil {
		SendError(w, "SERVICE_UNAVAILABLE", "Download service belum siap", http.StatusServiceUnavailable)
		return
	}

	job, err := downloadQueue.Submit(req.URL)
	if err != nil {
		SendError(w, "QUEUE_FULL", err.Error(), http.StatusServiceUnavailable)
		return
	}

	SendSuccess(w, map[string]interface{}{
		"job_id":  job.ID,
		"status":  job.Status,
		"message": "Job download telah dibuat. Gunakan endpoint /download/{job_id}/status untuk mengecek status.",
	}, http.StatusAccepted)
}

// DownloadStatusHandler menangani request status job
func DownloadStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Path[len("/download/"):]
	if idx := strings.Index(jobID, "/"); idx != -1 {
		jobID = jobID[:idx]
	}

	if jobID == "" {
		SendError(w, "MISSING_JOB_ID", "Job ID tidak ditemukan di URL", http.StatusBadRequest)
		return
	}

	if downloadQueue == nil {
		SendError(w, "SERVICE_UNAVAILABLE", "Download service belum siap", http.StatusServiceUnavailable)
		return
	}

	job, ok := downloadQueue.Get(jobID)
	if !ok {
		SendError(w, "JOB_NOT_FOUND", "Job tidak ditemukan", http.StatusNotFound)
		return
	}

	SendSuccess(w, job, http.StatusOK)
}

// DownloadFileHandler menangani download file yang sudah selesai
func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, "METHOD_NOT_ALLOWED", "Method tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	jobID := r.URL.Path[len("/download/"):]
	if idx := strings.Index(jobID, "/file"); idx != -1 {
		jobID = jobID[:idx]
	}

	if jobID == "" {
		SendError(w, "MISSING_JOB_ID", "Job ID tidak ditemukan di URL", http.StatusBadRequest)
		return
	}

	if downloadQueue == nil {
		SendError(w, "SERVICE_UNAVAILABLE", "Download service belum siap", http.StatusServiceUnavailable)
		return
	}

	job, ok := downloadQueue.Get(jobID)
	if !ok {
		SendError(w, "JOB_NOT_FOUND", "Job tidak ditemukan", http.StatusNotFound)
		return
	}

	if job.Status != jobs.StatusCompleted {
		SendError(w, "JOB_NOT_READY", "File belum siap untuk didownload", http.StatusBadRequest)
		return
	}

	// File ada di jobDir, serve langsung
	cfg := config.Load()
	jobDir := filepath.Join(cfg.TempDir, fmt.Sprintf("job_%s", job.ID))
	matches, _ := filepath.Glob(filepath.Join(jobDir, "*.mp3"))
	if len(matches) == 0 {
		SendError(w, "FILE_NOT_FOUND", "File tidak ditemukan", http.StatusNotFound)
		return
	}

	outputFile := matches[0]
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(outputFile)))
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeFile(w, r, outputFile)
}

func isValidURL(rawurl string) bool {
	u, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return false
	}

	if u.Scheme != "https" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	allowedHosts := []string{
		"youtube.com",
		"www.youtube.com",
		"m.youtube.com",
		"youtu.be",
		"soundcloud.com",
		"www.soundcloud.com",
	}

	for _, allowed := range allowedHosts {
		if host == allowed {
			return true
		}
		if strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}
