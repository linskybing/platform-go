package utils

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogAuditConcurrency - Test thread safety
func TestLogAuditConcurrency(t *testing.T) {
	repo := &mockAuditRepo{}

	// Log multiple times concurrently
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_ = LogAudit(
				strconv.Itoa(id), "127.0.0.1", "test", "action",
				"resource", "id_"+string(rune(id)),
				nil, nil, "test", repo,
			)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// All logs should be created (with mutex protection)
	repo.mu.Lock()
	actual := len(repo.createdLogs)
	repo.mu.Unlock()
	assert.Equal(t, numGoroutines, actual)
}
