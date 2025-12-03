package services_test

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/linskybing/platform-go/src/dto"
	"github.com/linskybing/platform-go/src/models"
	"github.com/linskybing/platform-go/src/repositories"
	"github.com/linskybing/platform-go/src/repositories/mock_repositories"
	"github.com/linskybing/platform-go/src/services"
	"github.com/linskybing/platform-go/src/utils"
)

func setupGroupMocks(t *testing.T) (*services.GroupService, *mock_repositories.MockGroupRepo, *mock_repositories.MockAuditRepo, *gin.Context) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockGroup := mock_repositories.NewMockGroupRepo(ctrl)
	mockAudit := mock_repositories.NewMockAuditRepo(ctrl)
	repos := &repositories.Repos{
		Group: mockGroup,
		Audit: mockAudit,
	}

	svc := services.NewGroupService(repos)
	c, _ := gin.CreateTestContext(nil)

	// Mock audit log globally
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repositories.AuditRepo) {
	}

	return svc, mockGroup, mockAudit, c
}

func TestGroupServiceCRUD(t *testing.T) {
	svc, mockGroup, _, c := setupGroupMocks(t)

	t.Run("ListGroups success", func(t *testing.T) {
		mockGroup.EXPECT().GetAllGroups().Return([]models.Group{
			{GID: 1, GroupName: "dev"},
			{GID: 2, GroupName: "ops"},
		}, nil)

		groups, err := svc.ListGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(groups))
		}
	})

	t.Run("GetGroup success", func(t *testing.T) {
		mockGroup.EXPECT().GetGroupByID(uint(1)).Return(models.Group{GID: 1, GroupName: "dev"}, nil)
		group, err := svc.GetGroup(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "dev" {
			t.Fatalf("expected dev, got %s", group.GroupName)
		}
	})

	t.Run("GetGroup not found", func(t *testing.T) {
		mockGroup.EXPECT().GetGroupByID(uint(99)).Return(models.Group{}, errors.New("not found"))
		_, err := svc.GetGroup(99)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("CreateGroup success", func(t *testing.T) {
		input := dto.GroupCreateDTO{GroupName: "qa", Description: nil}
		mockGroup.EXPECT().CreateGroup(gomock.Any()).Return(nil)

		group, err := svc.CreateGroup(c, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "qa" {
			t.Fatalf("expected qa, got %s", group.GroupName)
		}
	})

	t.Run("CreateGroup reserved name", func(t *testing.T) {
		input := dto.GroupCreateDTO{GroupName: "super"}
		_, err := svc.CreateGroup(c, input)
		if !errors.Is(err, services.ErrReservedGroupName) {
			t.Fatalf("expected reserved group name error, got %v", err)
		}
	})

	t.Run("UpdateGroup success", func(t *testing.T) {
		old := models.Group{GID: 1, GroupName: "dev"}
		mockGroup.EXPECT().GetGroupByID(uint(1)).Return(old, nil)
		mockGroup.EXPECT().UpdateGroup(gomock.Any()).Return(nil)

		newName := "devops"
		newDesc := "updated description"
		input := dto.GroupUpdateDTO{GroupName: &newName, Description: &newDesc}
		group, err := svc.UpdateGroup(c, 1, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "devops" || group.Description != "updated description" {
			t.Fatalf("unexpected updated group %+v", group)
		}
	})

	t.Run("UpdateGroup reserved name", func(t *testing.T) {
		old := models.Group{GID: 1, GroupName: "dev"}
		mockGroup.EXPECT().GetGroupByID(uint(1)).Return(old, nil)
		newName := "super"
		input := dto.GroupUpdateDTO{GroupName: &newName}
		_, err := svc.UpdateGroup(c, 1, input)
		if !errors.Is(err, services.ErrReservedGroupName) {
			t.Fatalf("expected reserved group name error, got %v", err)
		}
	})

	t.Run("UpdateGroup not found", func(t *testing.T) {
		mockGroup.EXPECT().GetGroupByID(uint(99)).Return(models.Group{}, errors.New("not found"))
		newName := "newname"
		input := dto.GroupUpdateDTO{GroupName: &newName}
		_, err := svc.UpdateGroup(c, 99, input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("DeleteGroup success", func(t *testing.T) {
		existing := models.Group{GID: 1, GroupName: "dev"}
		mockGroup.EXPECT().GetGroupByID(uint(1)).Return(existing, nil)
		mockGroup.EXPECT().DeleteGroup(uint(1)).Return(nil)

		err := svc.DeleteGroup(c, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("DeleteGroup not found", func(t *testing.T) {
		mockGroup.EXPECT().GetGroupByID(uint(99)).Return(models.Group{}, errors.New("not found"))
		err := svc.DeleteGroup(c, 99)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
