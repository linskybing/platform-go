package application_test

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/types"
)

func TestUpdateConfigFile_Success(t *testing.T) {
	svc, cfRepo, resRepo, auditRepo, _, _, _, c := setupConfigFileService(t)

	existingCommit := &configfile.ConfigCommit{ID: "1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "initial"}
	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return existingCommit, nil
	}
	newCommit := &configfile.ConfigCommit{ID: "2", ProjectID: "1", BlobHash: "hash2", AuthorID: "1", Message: "update"}
	cfRepo.store = func(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error) {
		return newCommit, nil
	}

	resRepo.createResource = func(ctx context.Context, res *resource.Resource) error {
		return nil
	}
	auditRepo.createAuditLog = func(a *audit.AuditLog) error {
		return nil
	}

	rawYaml := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: testpod"
	input := configfile.ConfigFileUpdateDTO{
		RawYaml: &rawYaml,
	}

	cf, err := svc.UpdateConfigFile(c.Request.Context(), "1", input, c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.ID != "2" {
		t.Fatalf("expected commit id 2, got %s", cf.ID)
	}
}
