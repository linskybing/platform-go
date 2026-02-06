package image

import (
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/image"
	"gorm.io/gorm"
)

// fakeRepo implements the minimal methods used by ApproveRequest
type fakeRepo struct {
	reqs    map[string]*image.ImageRequest
	created []*image.ImageAllowList
	repos   map[string]*image.ContainerRepository
	tags    map[string]*image.ContainerTag
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{reqs: make(map[string]*image.ImageRequest), repos: make(map[string]*image.ContainerRepository), tags: make(map[string]*image.ContainerTag)}
}

func (f *fakeRepo) CreateRequest(req *image.ImageRequest) error {
	f.reqs[req.ID] = req
	return nil
}

func (f *fakeRepo) FindRequestByID(id string) (*image.ImageRequest, error) {
	r, ok := f.reqs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}

func (f *fakeRepo) UpdateRequest(req *image.ImageRequest) error {
	f.reqs[req.ID] = req
	return nil
}

func (f *fakeRepo) CreateAllowListRule(rule *image.ImageAllowList) error {
	// Populate Repository and Tag objects from stored maps if possible
	if rule.RepositoryID != "" {
		if r, ok := f.repos[rule.RepositoryID]; ok {
			rule.Repository = r
		}
	}
	if rule.TagID != nil && *rule.TagID != "" {
		if tg, ok := f.tags[*rule.TagID]; ok {
			rule.Tag = tg
		}
	}
	f.created = append(f.created, rule)
	return nil
}

// Unused methods required by interface but not needed for this test
func (f *fakeRepo) ListRequests(projectID *string, status string) ([]image.ImageRequest, error) {
	return nil, nil
}
func (f *fakeRepo) FindAllRequests() ([]image.ImageRequest, error)                 { return nil, nil }
func (f *fakeRepo) FindRequestsByUserID(userID uint) ([]image.ImageRequest, error) { return nil, nil }
func (f *fakeRepo) ListAllowedImages(projectID *string) ([]image.ImageAllowList, error) {
	return nil, nil
}
func (f *fakeRepo) FindAllowListRule(projectID *string, repoFullName, tagName string) (*image.ImageAllowList, error) {
	return nil, nil
}
func (f *fakeRepo) FindOrCreateRepository(repo *image.ContainerRepository) error {
	if repo.ID == "" {
		repo.ID = "test-repo-1"
	}
	// store or update
	f.repos[repo.ID] = repo
	return nil
}
func (f *fakeRepo) FindOrCreateTag(tag *image.ContainerTag) error {
	if tag.ID == "" {
		tag.ID = "test-tag-1"
	}
	f.tags[tag.ID] = tag
	return nil
}
func (f *fakeRepo) CheckImageAllowed(projectID *string, repoFullName string, tagName string) (bool, error) {
	return false, nil
}
func (f *fakeRepo) DisableAllowListRule(id string) error                             { return nil }
func (f *fakeRepo) UpdateClusterStatus(status *image.ClusterImageStatus) error       { return nil }
func (f *fakeRepo) GetClusterStatus(tagID string) (*image.ClusterImageStatus, error) { return nil, nil }
func (f *fakeRepo) WithTx(tx *gorm.DB) image.Repository                              { return f }
func (f *fakeRepo) GetTagByDigest(repoID string, digest string) (*image.ContainerTag, error) {
	return nil, nil
}

func TestApproveRequest(t *testing.T) {
	repo := newFakeRepo()
	svc := NewImageService(repo)

	// prepare a request
	req := &image.ImageRequest{
		UserID:         "10",
		InputImageName: "myrepo/myimage",
		InputTag:       "v1",
		ProjectID:      nil,
		Status:         "pending",
	}
	// set ID and store in fake repo
	req.ID = "1"
	repo.reqs["1"] = req

	approver := "99"

	err := svc.ApproveRequest("1", "ok", false, approver)
	if err != nil {
		t.Fatalf("ApproveRequest returned error: %v", err)
	}

	// fetch updated request
	updatedReq, _ := repo.FindRequestByID("1")
	if updatedReq.Status != "approved" {
		t.Fatalf("expected request status approved, got %s", updatedReq.Status)
	}
	// verify allowed image created
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 allowed image created, got %d", len(repo.created))
	}
	ai := repo.created[0]
	if ai.Repository.FullName != req.InputImageName || ai.Tag.Name != req.InputTag {
		t.Fatalf("allowed image mismatch: %+v", ai)
	}
	if ai.RequestID == nil || *ai.RequestID != req.ID {
		t.Fatalf("request id on allowlist not set: %v", ai.RequestID)
	}
	if ai.CreatedBy != approver {
		t.Fatalf("created_by not recorded: %v", ai.CreatedBy)
	}
	if !ai.IsEnabled {
		t.Fatalf("allowlist rule not enabled")
	}
}
