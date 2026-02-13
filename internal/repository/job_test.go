package repository

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, JobRepo) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&job.Job{})
	require.NoError(t, err)

	return db, NewJobRepo(db)
}

func TestJobRepo_Create(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	j := &job.Job{
		ID:        "job-1",
		ProjectID: "proj-1",
		UserID:    "user-1",
		Status:    "submitted",
	}

	err := repo.Create(ctx, j)
	require.NoError(t, err)

	found, err := repo.Get(ctx, "job-1")
	require.NoError(t, err)
	assert.Equal(t, j.ID, found.ID)
	assert.Equal(t, j.ProjectID, found.ProjectID)
}

func TestJobRepo_Get_NotFound(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.Get(ctx, "non-existent")
	assert.Error(t, err)
	assert.True(t, gorm.ErrRecordNotFound == err || err.Error() == "record not found")
}

func TestJobRepo_UpdateStatus(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	j := &job.Job{
		ID:     "job-1",
		Status: "submitted",
	}
	require.NoError(t, repo.Create(ctx, j))

	err := repo.UpdateStatus(ctx, "job-1", "running", nil)
	require.NoError(t, err)

	found, err := repo.Get(ctx, "job-1")
	require.NoError(t, err)
	assert.Equal(t, "running", found.Status)
	assert.Empty(t, found.ErrorMessage)

	errMsg := "something failed"
	err = repo.UpdateStatus(ctx, "job-1", "failed", &errMsg)
	require.NoError(t, err)

	found, err = repo.Get(ctx, "job-1")
	require.NoError(t, err)
	assert.Equal(t, "failed", found.Status)
	assert.Equal(t, "something failed", found.ErrorMessage)
}

func TestJobRepo_ListByProject(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-1", ProjectID: "p1", Status: "running"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-2", ProjectID: "p1", Status: "completed"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-3", ProjectID: "p2", Status: "submitted"}))

	jobs, err := repo.ListByProject(ctx, "p1")
	require.NoError(t, err)
	assert.Len(t, jobs, 2)
}

func TestJobRepo_ListByUser(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-1", UserID: "u1"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-2", UserID: "u1"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-3", UserID: "u2"}))

	jobs, err := repo.ListByUser(ctx, "u1")
	require.NoError(t, err)
	assert.Len(t, jobs, 2)
}

func TestJobRepo_ListByStatus(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-1", Status: "pending"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-2", Status: "running"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-3", Status: "completed"}))

	jobs, err := repo.ListByStatus(ctx, []string{"pending", "running"})
	require.NoError(t, err)
	assert.Len(t, jobs, 2)
}

func TestJobRepo_ListByProjectAndStatuses(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-1", ProjectID: "p1", Status: "failed"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-2", ProjectID: "p1", Status: "running"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "job-3", ProjectID: "p2", Status: "running"}))

	jobs, err := repo.ListByProjectAndStatuses(ctx, "p1", []string{"failed"})
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "job-1", jobs[0].ID)
}

func TestJobRepo_CountByUserProjectAndStatuses(t *testing.T) {
	_, repo := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &job.Job{ID: "j1", UserID: "u1", ProjectID: "p1", Status: "running"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "j2", UserID: "u1", ProjectID: "p1", Status: "queued"}))
	require.NoError(t, repo.Create(ctx, &job.Job{ID: "j3", UserID: "u1", ProjectID: "p1", Status: "completed"}))

	count, err := repo.CountByUserProjectAndStatuses(ctx, "u1", "p1", []string{"running", "queued"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func BenchmarkJobRepo_Create(b *testing.B) {
	_, repo := setupTestDB(&testing.T{})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j := &job.Job{
			ID:     "job-bench", // overwrite same ID
			Status: "pending",
		}
		_ = repo.Create(ctx, j)
	}
}
