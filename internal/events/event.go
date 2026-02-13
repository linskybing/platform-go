package events

import (
"time"
)

type EventType string

const (
// Job Events
JobCreated   EventType = "job.created"
JobCompleted EventType = "job.completed"
JobCancelled EventType = "job.cancelled"
JobFailed    EventType = "job.failed"

// User Events
UserCreated EventType = "user.created"
UserUpdated EventType = "user.updated"
UserDeleted EventType = "user.deleted"

// Project Events
ProjectCreated EventType = "project.created"
ProjectUpdated EventType = "project.updated"
ProjectDeleted EventType = "project.deleted"

// Storage Events
StorageInitialized EventType = "storage.initialized"
StorageDeleted     EventType = "storage.deleted"

// Form Events
FormCreated       EventType = "form.created"
FormStatusChanged EventType = "form.status_changed"
)

type Event struct {
ID        string      `json:"id"`
Type      EventType   `json:"type"`
Payload   interface{} `json:"payload"`
Timestamp time.Time   `json:"timestamp"`
UserID    string      `json:"user_id,omitempty"`
}
