package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Youtube_donwloader/config"
)

var (
	ytdlpPathOnce sync.Once
	ytdlpPath     string
)

func getYtdlpPath() string {
	ytdlpPathOnce.Do(func() {
		path := os.Getenv("YTDLP_PATH")
		if path == "" {
			if _, err := os.Stat("yt-dlp.exe"); err == nil {
				path = "yt-dlp.exe"
			} else {
				path = "yt-dlp"
			}
		}

		if !filepath.IsAbs(path) && !strings.Contains(path, string(os.PathSeparator)) {
			if absPath, err := filepath.Abs(path); err == nil {
				if _, err := os.Stat(absPath); err == nil {
					path = absPath
				} else {
					if _, err := os.Stat("." + string(os.PathSeparator) + path); err == nil {
						path, _ = filepath.Abs("." + string(os.PathSeparator) + path)
					}
				}
			}
		}
		ytdlpPath = path
	})
	return ytdlpPath
}

// ProcessDownload memproses job download
func ProcessDownload(job *DownloadJob) error {
	cfg := config.Load()
	baseTempDir := cfg.TempDir
	jobDir := filepath.Join(baseTempDir, fmt.Sprintf("job_%s", job.ID))

	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori temporary: %v", err)
	}

	fileName := filepath.Join(jobDir, "output")

	options := [][]string{
		{
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--output", fileName + ".%(ext)s",
			"--no-playlist",
			"--format", "bestaudio/best",
			"--extractor-args", "youtube:player_client=android",
		},
		{
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--output", fileName + ".%(ext)s",
			"--no-playlist",
			"--format", "bestaudio/best",
			"--extractor-args", "youtube:player_client=ios",
		},
		{
			"--extract-audio",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--output", fileName + ".%(ext)s",
			"--no-playlist",
			"--format", "bestaudio/best",
		},
	}

	ytdlpPath := getYtdlpPath()

	for i, opts := range options {
		args := append(opts, job.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		cmd := exec.CommandContext(ctx, ytdlpPath, args...)

		cmd.Stdout = io.Discard
		stderrBuf := newLimitedBuffer(4096)
		cmd.Stderr = stderrBuf

		start := time.Now()
		err := cmd.Run()
		duration := time.Since(start)
		cancel()

		if err != nil {
			exitCode := 0
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			}

			log.Printf(
				"[processor] job=%s attempt=%d exitCode=%d duration_ms=%d stderr=%q err=%v",
				job.ID, i+1, exitCode, duration.Milliseconds(), stderrBuf.String(), err,
			)

			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("download timeout setelah %s", duration)
			}

			if i < len(options)-1 {
				continue
			}

			return fmt.Errorf("gagal mendownload setelah mencoba %d metode", len(options))
		}

		matches, _ := filepath.Glob(filepath.Join(jobDir, "*.mp3"))
		if len(matches) == 0 {
			return fmt.Errorf("file MP3 tidak ditemukan setelah download")
		}

		outputFile := matches[0]
		job.FileURL = fmt.Sprintf("/download/%s/file", job.ID)

		log.Printf("[processor] job=%s completed, file=%s", job.ID, outputFile)
		return nil
	}

	return fmt.Errorf("semua metode download gagal")
}

type limitedBuffer struct {
	buf bytes.Buffer
	max int
}

func newLimitedBuffer(max int) *limitedBuffer {
	return &limitedBuffer{max: max}
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	remaining := l.max - l.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	return l.buf.Write(p)
}

func (l *limitedBuffer) String() string {
	return l.buf.String()
}
