package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/repositories/mock_repositories"
	"github.com/linskybing/platform-go/utils"
)

func setupResourceMocks(t *testing.T) (*ResourceService,
	*mock_repositories.MockResourceRepo,
	*mock_repositories.MockAuditRepo,
	*gin.Context) {

	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockResource := mock_repositories.NewMockResourceRepo(ctrl)
	mockAudit := mock_repositories.NewMockAuditRepo(ctrl)

	repos := &repositories.Repos{
		Resource: mockResource,
		Audit:    mockAudit,
	}

	svc := NewResourceService(repos)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repositories.AuditRepo) {
	}
	utils.GetUserIDFromContext = func(c *gin.Context) (uint, error) { return 1, nil }

	return svc, mockResource, mockAudit, c
}

func TestResourceCRUD(t *testing.T) {
	svc, mockResource, _, c := setupResourceMocks(t)

	// ----- Create -----
	resource := &models.Resource{
		RID:  1,
		Type: "Pod",
		Name: "res1",
	}
	mockResource.EXPECT().CreateResource(resource).Return(nil)

	created, err := svc.CreateResource(c, resource)
	if err != nil || created.RID != 1 {
		t.Fatalf("CreateResource failed: %v", err)
	}

	// ----- Get -----
	mockResource.EXPECT().GetResourceByID(uint(1)).Return(resource, nil)
	got, err := svc.GetResource(1)
	if err != nil || got.RID != 1 {
		t.Fatalf("GetResource failed: %v", err)
	}

	// ----- Update -----
	newName := "res2"
	updateDTO := dto.ResourceUpdateDTO{Name: &newName}
	mockResource.EXPECT().GetResourceByID(uint(1)).Return(resource, nil)
	mockResource.EXPECT().UpdateResource(resource).Return(nil)

	updated, err := svc.UpdateResource(c, 1, updateDTO)
	if err != nil || updated.Name != "res2" {
		t.Fatalf("UpdateResource failed: %v", err)
	}

	// ----- Delete -----
	mockResource.EXPECT().GetResourceByID(uint(1)).Return(resource, nil)
	mockResource.EXPECT().DeleteResource(uint(1)).Return(nil)

	if err := svc.DeleteResource(c, 1); err != nil {
		t.Fatalf("DeleteResource failed: %v", err)
	}
}
