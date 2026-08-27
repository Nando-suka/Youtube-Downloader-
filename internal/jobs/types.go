package jobs

import (
	"time"
)

// JobStatus merepresentasikan status dari sebuah job
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

// DownloadJob merepresentasikan job download
type DownloadJob struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Status    JobStatus `json:"status"`
	Progress  int       `json:"progress"` // 0-100
	Error     string    `json:"error,omitempty"`
	FileURL   string    `json:"file_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
