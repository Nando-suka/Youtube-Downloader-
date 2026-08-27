package jobs

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Queue mengelola job queue untuk download
type Queue struct {
	jobs         map[string]*DownloadJob
	mutex        sync.RWMutex
	workers      int
	jobChan      chan *DownloadJob
	processor    func(*DownloadJob) error
	onJobDone    func(*DownloadJob) // Callback ketika job selesai
}

// NewQueue membuat queue baru dengan worker pool
func NewQueue(workers int, processor func(*DownloadJob) error) *Queue {
	q := &Queue{
		jobs:      make(map[string]*DownloadJob),
		workers:   workers,
		jobChan:   make(chan *DownloadJob, 100),
		processor: processor,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		go q.worker(i)
	}

	// Cleanup goroutine untuk menghapus job lama
	go q.cleanup()

	return q
}

// SetOnJobDone mengatur callback yang dipanggil ketika job selesai
func (q *Queue) SetOnJobDone(callback func(*DownloadJob)) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.onJobDone = callback
}

// Submit menambahkan job baru ke queue
func (q *Queue) Submit(url string) (*DownloadJob, error) {
	now := time.Now()
	job := &DownloadJob{
		ID:        uuid.New().String(),
		URL:       url,
		Status:    StatusPending,
		Progress:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	q.mutex.Lock()
	q.jobs[job.ID] = job
	q.mutex.Unlock()

	// Kirim ke channel untuk diproses
	select {
	case q.jobChan <- job:
		log.Printf("[queue] Job %s submitted for URL: %s", job.ID, url)
	default:
		return nil, fmt.Errorf("queue is full, please try again later")
	}

	return job, nil
}

// Get mengambil job berdasarkan ID
func (q *Queue) Get(id string) (*DownloadJob, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	job, ok := q.jobs[id]
	if !ok {
		return nil, false
	}
	// Return copy untuk thread safety
	jobCopy := *job
	return &jobCopy, true
}

// worker memproses job dari channel
func (q *Queue) worker(id int) {
	for job := range q.jobChan {
		log.Printf("[worker-%d] Processing job %s", id, job.ID)

		q.mutex.Lock()
		job.Status = StatusProcessing
		now := time.Now()
		job.StartedAt = &now
		job.UpdatedAt = now
		q.mutex.Unlock()

		// Proses job
		err := q.processor(job)

		q.mutex.Lock()
		finishedAt := time.Now()
		job.FinishedAt = &finishedAt
		job.UpdatedAt = finishedAt

		if err != nil {
			job.Status = StatusFailed
			job.Error = err.Error()
			log.Printf("[worker-%d] Job %s failed: %v", id, job.ID, err)
		} else {
			job.Status = StatusCompleted
			job.Progress = 100
			log.Printf("[worker-%d] Job %s completed", id, job.ID)
		}
		q.mutex.Unlock()

		// Call callback jika ada
		if q.onJobDone != nil {
			jobCopy := *job
			q.onJobDone(&jobCopy)
		}
	}
}

// cleanup menghapus job yang sudah lama (lebih dari 1 jam setelah selesai)
func (q *Queue) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		q.mutex.Lock()
		now := time.Now()
		for id, job := range q.jobs {
			if job.FinishedAt != nil {
				if now.Sub(*job.FinishedAt) > time.Hour {
					delete(q.jobs, id)
					log.Printf("[cleanup] Removed old job %s", id)
				}
			}
		}
		q.mutex.Unlock()
	}
}
