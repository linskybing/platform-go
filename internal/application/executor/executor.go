package executor

import "context"

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusSubmitted JobStatus = "submitted"
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

type SubmitType string

const (
	SubmitTypeJob      SubmitType = "job"
	SubmitTypeWorkflow SubmitType = "workflow"
)

// SubmitRequest contains all information needed to submit a job
type SubmitRequest struct {
	JobID          string
	ConfigCommitID string
	ProjectID      string
	Namespace      string
	UserID         string
	Username       string
	Resources      []ResourcePayload
	SubmitType     SubmitType
	QueueName      string // for future scheduler
	Priority       int32  // for future scheduler
}

// ResourcePayload represents a Kubernetes resource to be deployed
type ResourcePayload struct {
	Name     string
	Kind     string
	JSONData []byte
}

// SubmitResult contains the result of a job submission
type SubmitResult struct {
	JobID  string
	Status JobStatus
}

// Executor defines the interface for job execution
// This abstraction allows switching between local K8s deployment and external schedulers
type Executor interface {
	// Submit submits a job for execution
	Submit(ctx context.Context, req *SubmitRequest) (*SubmitResult, error)

	// Cancel cancels a running or queued job
	Cancel(ctx context.Context, jobID string) error

	// Status retrieves the current status of a job
	Status(ctx context.Context, jobID string) (JobStatus, error)
}
