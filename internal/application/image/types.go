package image

import (
	"sync"
	"time"
)

func ptrTime(t time.Time) *time.Time { return &t }

type PullJobStatus struct {
	JobID     string    `json:"job_id"`
	ImageName string    `json:"image_name"`
	ImageTag  string    `json:"image_tag"`
	Status    string    `json:"status"`
	Progress  int       `json:"progress"`
	Message   string    `json:"message"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PullJobTracker struct {
	mu         sync.RWMutex
	jobs       map[string]*PullJobStatus
	chans      map[string][]chan *PullJobStatus
	failedJobs []*PullJobStatus
	maxHistory int
}

var pullTracker = &PullJobTracker{
	jobs:       make(map[string]*PullJobStatus),
	chans:      make(map[string][]chan *PullJobStatus),
	failedJobs: make([]*PullJobStatus, 0),
	maxHistory: 50,
}

func (pt *PullJobTracker) AddJob(jobID, imageName, imageTag string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.jobs[jobID] = &PullJobStatus{
		JobID:     jobID,
		ImageName: imageName,
		ImageTag:  imageTag,
		Status:    "pending",
		Progress:  0,
		UpdatedAt: time.Now(),
	}
}

func (pt *PullJobTracker) GetJob(jobID string) *PullJobStatus {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.jobs[jobID]
}

func (pt *PullJobTracker) UpdateJob(jobID string, status string, progress int, message string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if job, ok := pt.jobs[jobID]; ok {
		job.Status = status
		job.Progress = progress
		job.Message = message
		job.UpdatedAt = time.Now()

		if chans, ok := pt.chans[jobID]; ok {
			for _, ch := range chans {
				select {
				case ch <- job:
				default:
				}
			}
		}
	}
}

func (pt *PullJobTracker) Subscribe(jobID string) <-chan *PullJobStatus {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	ch := make(chan *PullJobStatus, 10)
	pt.chans[jobID] = append(pt.chans[jobID], ch)
	return ch
}

func (pt *PullJobTracker) RemoveJob(jobID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	if job, ok := pt.jobs[jobID]; ok && job.Status == "failed" {
		pt.failedJobs = append(pt.failedJobs, job)
		if len(pt.failedJobs) > pt.maxHistory {
			pt.failedJobs = pt.failedJobs[len(pt.failedJobs)-pt.maxHistory:]
		}
	}

	delete(pt.jobs, jobID)
}

func (pt *PullJobTracker) GetFailedJobs(limit int) []*PullJobStatus {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if limit <= 0 || limit > len(pt.failedJobs) {
		limit = len(pt.failedJobs)
	}

	result := make([]*PullJobStatus, 0, limit)
	for i := len(pt.failedJobs) - 1; i >= len(pt.failedJobs)-limit && i >= 0; i-- {
		result = append(result, pt.failedJobs[i])
	}
	return result
}

func (pt *PullJobTracker) GetActiveJobs() []*PullJobStatus {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	result := make([]*PullJobStatus, 0, len(pt.jobs))
	for _, job := range pt.jobs {
		result = append(result, job)
	}
	return result
}
