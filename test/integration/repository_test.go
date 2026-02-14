//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestRepository_Integration(t *testing.T) {
	// Ensure DB is initialized
	testCtx := GetTestContext()

	repos := repository.NewRepositories(db.DB)
	ctx := context.Background()

	t.Run("JobRepo_DeepPagination_Performance", func(t *testing.T) {
		// Just a smoke test for the repo connection, benchmarks are separate
		jobID := "00000000-0000-0000-0000-000000000006"
		err := repos.Job.Create(ctx, &job.Job{
			ID:             jobID,
			ProjectID:      testCtx.TestProject.ID,
			UserID:         testCtx.TestUser.ID,
			ConfigCommitID: "00000000-0000-0000-0000-000000000000",
			Status:         "submitted",
		})
		require.NoError(t, err)

		found, err := repos.Job.Get(ctx, jobID)
		require.NoError(t, err)
		assert.Equal(t, "submitted", found.Status)
	})

	t.Run("UserGroupRepo_Preload_Verify", func(t *testing.T) {
		// Validates that Preload("Group") is working (N+1 fix)
		// We'll inspect if the returned UserGroup has the Group struct populated

		testCtx := GetTestContext()

		// Use repo to fetch
		ugs, err := repos.UserGroup.GetUserGroupsByUID(ctx, testCtx.TestUser.ID)
		require.NoError(t, err)
		require.NotEmpty(t, ugs)

		// Verify Preload
		assert.NotEmpty(t, ugs[0].Group.Name, "Group should be preloaded")
	})

	t.Run("StorageRepo_Preload_Verify", func(t *testing.T) {
		// Create a user storage
		testCtx := GetTestContext()

		s := storage.Storage{
			ID:             "00000000-0000-0000-0000-000000000007",
			OwnerID:        testCtx.TestUser.ID,
			Name:           "test-storage",
			K8sNamespace:   "test-ns",
			PVCName:        "test-pvc",
			Capacity:       1, // 1Gi (int)
			AffinityConfig: datatypes.JSON([]byte("{}")),
		}

		err := db.DB.Create(&s).Error
		require.NoError(t, err)
		defer db.DB.Delete(&s)

		// Fetch via repo
		storages, err := repos.Storage.ListStorageByOwnerID(ctx, testCtx.TestUser.ID)
		require.NoError(t, err)
		require.NotEmpty(t, storages)

		found := false
		for _, st := range storages {
			if st.ID == "00000000-0000-0000-0000-000000000007" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}
