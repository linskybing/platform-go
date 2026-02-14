//go:build integration

package integration

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobPlugin_Integration(t *testing.T) {
	ctx := GetTestContext()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() { _ = cleaner.Cleanup() })

	repos := repository.NewRepositories(db.DB)

	// Cleanup manual jobs
	t.Cleanup(func() {
		db.DB.Exec("DELETE FROM jobs WHERE CAST(id AS text) LIKE '00000000-0000-0000-0000-%'")
	})

	t.Run("CreateJobViaRepo", func(t *testing.T) {
		jobID := "00000000-0000-0000-0000-000000000001"
		j := &job.Job{
			ID:             jobID,
			ProjectID:      ctx.TestProject.ID,
			UserID:         ctx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "submitted",
			SubmitType:     "job",
			QueueName:      "default",
			CreatedAt:      time.Now(),
		}

		// Use a dummy request to get a context, or background
		// repos methods use gorm WithContext.
		req, _ := http.NewRequest("GET", "/", nil)
		err := repos.Job.Create(req.Context(), j)
		require.NoError(t, err)

		// Verify
		var found job.Job
		err = db.DB.First(&found, "id = ?", jobID).Error
		assert.NoError(t, err)
		assert.Equal(t, "submitted", found.Status)
	})

	t.Run("ListJobs_Success", func(t *testing.T) {
		// Populate logic
		j1 := &job.Job{
			ID:             "00000000-0000-0000-0000-000000000002",
			ProjectID:      ctx.TestProject.ID,
			UserID:         ctx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "running",
			CreatedAt:      time.Now(),
		}
		j2 := &job.Job{
			ID:             "00000000-0000-0000-0000-000000000003",
			ProjectID:      ctx.TestProject.ID,
			UserID:         ctx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "completed",
			CreatedAt:      time.Now(),
		}
		require.NoError(t, db.DB.Create(j1).Error)
		require.NoError(t, db.DB.Create(j2).Error)

		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err := client.GET("/jobs")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetJob_Success", func(t *testing.T) {
		jobID := "00000000-0000-0000-0000-000000000004"
		j := &job.Job{
			ID:             jobID,
			ProjectID:      ctx.TestProject.ID,
			UserID:         ctx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "running",
			CreatedAt:      time.Now(),
		}
		require.NoError(t, db.DB.Create(j).Error)

		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err := client.GET(fmt.Sprintf("/jobs/%s", jobID))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetJob_NotFound", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err := client.GET("/jobs/00000000-0000-0000-0000-ffffffffffff")
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SubmitJob_Success", func(t *testing.T) {
		// First create a config commit to refer to
		jobYaml := `apiVersion: batch/v1
kind: Job
metadata:
  name: submit-test-job
spec:
  template:
    spec:
      containers:
      - name: test
        image: nginx:latest
      restartPolicy: Never`
		createdConfig := createConfigFileAsManager(t, ctx, jobYaml, "for submission")
		
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		payload := map[string]interface{}{
			"project_id":       ctx.TestProject.ID,
			"config_commit_id": createdConfig.Commit.ID,
		}

		resp, err := client.POST("/jobs/submit", payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("CancelJob_Success", func(t *testing.T) {
		jobID := "00000000-0000-0000-0000-000000000005"
		j := &job.Job{
			ID:             jobID,
			ProjectID:      ctx.TestProject.ID,
			UserID:         ctx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "running",
			CreatedAt:      time.Now(),
		}
		require.NoError(t, db.DB.Create(j).Error)

		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err := client.POST(fmt.Sprintf("/jobs/%s/cancel", j.ID), nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updated job.Job
		db.DB.First(&updated, "id = ?", j.ID)
		assert.Equal(t, "cancelled", updated.Status)
	})
}

func randomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}
