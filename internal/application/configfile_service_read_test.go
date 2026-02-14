package application_test

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/configfile"
)

func TestConfigFileRead(t *testing.T) {
	svc, cfRepo, _, _, _, _, _, _ := setupConfigFileService(t)

	t.Run("ListConfigFiles", func(t *testing.T) {
		cfs := []configfile.ConfigCommit{{ID: "1", ProjectID: "10", BlobHash: "hash", AuthorID: "1", Message: "m"}}
		cfRepo.listAllCommits = func(ctx context.Context) ([]configfile.ConfigCommit, error) {
			return cfs, nil
		}

		res, err := svc.ListConfigFiles()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 config commit, got %d", len(res))
		}
	})

	t.Run("GetConfigFile", func(t *testing.T) {
		cf := &configfile.ConfigCommit{ID: "1", ProjectID: "10", BlobHash: "hash", AuthorID: "1", Message: "m"}
		cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
			return cf, nil
		}

		res, err := svc.GetConfigFile("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ID != "1" {
			t.Fatalf("expected commit id 1, got %s", res.ID)
		}
	})

	t.Run("ListConfigFilesByProjectID", func(t *testing.T) {
		cfs := []configfile.ConfigCommit{{ID: "1", ProjectID: "10", BlobHash: "hash", AuthorID: "1", Message: "m"}}
		cfRepo.getHistory = func(ctx context.Context, projectID string) ([]configfile.ConfigCommit, error) {
			return cfs, nil
		}

		res, err := svc.ListConfigFilesByProjectID("10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 config commit, got %d", len(res))
		}
	})
}
