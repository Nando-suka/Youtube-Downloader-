package handlers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.TrimSpace(r.FormValue("url"))
	if url == "" {
		http.Error(w, "URL tidak boleh kosong", http.StatusBadRequest)
		return
	}

	if !isValidURL(url) {
		http.Error(w, "URL tidak valid", http.StatusBadRequest)
		return
	}

	// Gunakan path yt-dlp dari environment variable atau default
	ytdlpPath := os.Getenv("YTDLP_PATH")
	if ytdlpPath == "" {
		// Default untuk Windows: coba yt-dlp.exe, jika tidak ada coba yt-dlp di PATH
		if _, err := os.Stat("yt-dlp.exe"); err == nil {
			ytdlpPath = "yt-dlp.exe"
		} else {
			ytdlpPath = "yt-dlp" // default to yt-dlp di PATH
		}
	}

	// Untuk Windows: jika path relatif dan tidak ada slash, konversi ke absolute path
	// atau gunakan format .\executable.exe
	if !filepath.IsAbs(ytdlpPath) && !strings.Contains(ytdlpPath, string(os.PathSeparator)) {
		// Coba resolve sebagai executable di current directory
		if absPath, err := filepath.Abs(ytdlpPath); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				ytdlpPath = absPath
			} else {
				// Jika tidak ditemukan, coba dengan .\ untuk Windows
				if _, err := os.Stat("." + string(os.PathSeparator) + ytdlpPath); err == nil {
					ytdlpPath, _ = filepath.Abs("." + string(os.PathSeparator) + ytdlpPath)
				}
			}
		}
	}

	fileName := filepath.Join("./tmp", fmt.Sprintf("output_%d", time.Now().UnixNano()))

	cmd := exec.Command(
		ytdlpPath,
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--output", fileName+".%(ext)s",
		"--no-playlist",
		url,
	)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Gagal mendownload: %v\nDetail Error: %s", err, stderr.String()),
			http.StatusInternalServerError,
		)
		return
	}

	matches, _ := filepath.Glob(fileName + ".mp3")
	if len(matches) == 0 {
		http.Error(w, "File MP3 tidak ditemukan", http.StatusInternalServerError)
		return
	}
	outputFile := matches[0]

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(outputFile)))
	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeFile(w, r, outputFile)

	defer os.Remove(outputFile)
}

func isValidURL(url string) bool {
	allowedDomains := []string{
		"youtube.com",
		"youtu.be",
		"soundcloud.com",
	}

	for _, domain := range allowedDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}
