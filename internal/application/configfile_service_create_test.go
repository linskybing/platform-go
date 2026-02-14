package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
)

func TestCreateConfigFile_Success(t *testing.T) {
	svc, cfRepo, resRepo, auditRepo, _, _, _, c := setupConfigFileService(t)

	commit := &configfile.ConfigCommit{ID: "c1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "initial commit"}
	cfRepo.store = func(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error) {
		return commit, nil
	}
	resRepo.createResource = func(ctx context.Context, res *resource.Resource) error {
		return nil
	}
	auditRepo.createAuditLog = func(a *audit.AuditLog) error {
		return nil
	}

	input := configfile.CreateConfigFileInput{
		RawYaml:   "apiVersion: v1\nkind: Pod\nmetadata:\n  name: testpod",
		ProjectID: "1",
	}

	cf, err := svc.CreateConfigFile(c.Request.Context(), input, c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.ID != "c1" {
		t.Fatalf("expected commit id c1, got %s", cf.ID)
	}
}

func TestCreateConfigFile_NoYAMLDocuments(t *testing.T) {
	svc, _, _, _, _, _, _, c := setupConfigFileService(t)

	utils.SplitYAMLDocuments = func(yamlStr string) []string { return []string{} }

	input := configfile.CreateConfigFileInput{
		RawYaml:   "",
		ProjectID: "1",
	}

	_, err := svc.CreateConfigFile(c.Request.Context(), input, c.MustGet("claims").(*types.Claims))
	if !errors.Is(err, application.ErrNoValidYAMLDocument) {
		t.Fatalf("expected ErrNoValidYAMLDocument, got %v", err)
	}
}
