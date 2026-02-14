package application_test

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/pkg/types"
)

func TestDeleteConfigFile_Success(t *testing.T) {
	svc, cfRepo, resRepo, auditRepo, userRepo, _, _, c := setupConfigFileService(t)

	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return &configfile.ConfigCommit{
			ID: "1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "test",
		}, nil
	}

	resRepo.listResourcesByCommitID = func(ctx context.Context, commitID string) ([]resource.Resource, error) {
		return []resource.Resource{{RID: "10", Name: "res1"}}, nil
	}

	userRepo.listUsersByProjectID = func(ctx context.Context, pid string) ([]user.User, error) {
		return []user.User{{Username: "user1"}}, nil
	}

	resRepo.deleteResource = func(ctx context.Context, rid string) error {
		return nil
	}
	cfRepo.deleteCommit = func(ctx context.Context, id string) error {
		return nil
	}
	auditRepo.createAuditLog = func(a *audit.AuditLog) error {
		return nil
	}

	err := svc.DeleteConfigFile(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteConfigFileInstance_Success(t *testing.T) {
	svc, cfRepo, resRepo, _, userRepo, _, _, _ := setupConfigFileService(t)

	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return &configfile.ConfigCommit{
			ID:        "1",
			ProjectID: "1",
			BlobHash:  "hash",
			AuthorID:  "1",
			Message:   "test",
		}, nil
	}

	resRepo.listResourcesByCommitID = func(ctx context.Context, commitID string) ([]resource.Resource, error) {
		return []resource.Resource{{RID: "1", Name: "res1"}}, nil
	}

	userRepo.listUsersByProjectID = func(ctx context.Context, pid string) ([]user.User, error) {
		return []user.User{{Username: "user1"}}, nil
	}

	err := svc.DeleteConfigFileInstance("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
