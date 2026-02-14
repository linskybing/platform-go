package application_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/utils"
)

func setupUserGroupService(t *testing.T) (*application.UserGroupService, *stubUserGroupRepoUG, *stubUserRepoLite, *stubProjectRepoLite, *stubGroupRepoLite, *gin.Context) {
	t.Helper()
	ugRepo := &stubUserGroupRepoUG{}
	userRepo := &stubUserRepoLite{}
	projectRepo := &stubProjectRepoLite{}
	groupRepo := &stubGroupRepoLite{}

	repos := &repository.Repos{UserGroup: ugRepo, User: userRepo, Project: projectRepo, Group: groupRepo}
	svc := application.NewUserGroupService(repos)
	ctx, _ := gin.CreateTestContext(nil)
	utils.LogAuditWithConsole = func(ctx *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}
	return svc, ugRepo, userRepo, projectRepo, groupRepo, ctx
}
