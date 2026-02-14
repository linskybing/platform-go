package application_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/pkg/types"
	"gorm.io/datatypes"
)

func TestCreateInstance_Success(t *testing.T) {
	svc, cfRepo, resRepo, _, _, projectRepo, userGroupRepo, c := setupConfigFileService(t)

	resRepo.listResourcesByCommitID = func(ctx context.Context, commitID string) ([]resource.Resource, error) {
		return []resource.Resource{{RID: "1", Type: resource.ResourceJob, ParsedYAML: datatypes.JSON([]byte("{}"))}}, nil
	}
	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return &configfile.ConfigCommit{ID: "1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "test"}, nil
	}
	projectRepo.getProjectByID = func(ctx context.Context, id string) (*project.Project, error) {
		return &project.Project{PID: "1", GID: "10"}, nil
	}
	userGroupRepo.getUserGroup = func(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
		return &group.UserGroup{UID: "1", GID: "10", Role: "admin"}, nil
	}

	err := svc.CreateInstance(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateInstance_UsesBlobFallback(t *testing.T) {
	svc, cfRepo, resRepo, _, _, projectRepo, userGroupRepo, c := setupConfigFileService(t)

	resRepo.listResourcesByCommitID = func(ctx context.Context, commitID string) ([]resource.Resource, error) {
		return []resource.Resource{}, nil
	}
	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return &configfile.ConfigCommit{ID: "1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "test"}, nil
	}
	blobPayload, err := json.Marshal("apiVersion: v1\nkind: Pod\nmetadata:\n  name: testpod")
	if err != nil {
		t.Fatalf("failed to build blob payload: %v", err)
	}
	cfRepo.getBlob = func(ctx context.Context, hash string) (*configfile.ConfigBlob, error) {
		return &configfile.ConfigBlob{Hash: "hash", Content: datatypes.JSON(blobPayload)}, nil
	}
	projectRepo.getProjectByID = func(ctx context.Context, id string) (*project.Project, error) {
		return &project.Project{PID: "1", GID: "10"}, nil
	}
	userGroupRepo.getUserGroup = func(ctx context.Context, uid, gid string) (*group.UserGroup, error) {
		return &group.UserGroup{UID: "1", GID: "10", Role: "admin"}, nil
	}

	err = svc.CreateInstance(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteInstance_Success(t *testing.T) {
	svc, cfRepo, resRepo, _, _, _, _, c := setupConfigFileService(t)

	resRepo.listResourcesByCommitID = func(ctx context.Context, commitID string) ([]resource.Resource, error) {
		return []resource.Resource{{RID: "1", ParsedYAML: datatypes.JSON([]byte("{}"))}}, nil
	}
	cfRepo.getCommit = func(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
		return &configfile.ConfigCommit{ID: "1", ProjectID: "1", BlobHash: "hash", AuthorID: "1", Message: "test"}, nil
	}

	err := svc.DeleteInstance(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
