package application_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupConfigFileService(t *testing.T) (*application.ConfigFileService, *stubConfigFileRepo,
	*stubResourceRepo, *stubAuditRepo,
	*stubUserRepo, *stubProjectRepo, *stubUserGroupRepo, *gin.Context) {

	t.Helper()
	cfRepo := &stubConfigFileRepo{}
	resRepo := &stubResourceRepo{}
	auditRepo := &stubAuditRepo{}
	projectRepo := &stubProjectRepo{}
	userGroupRepo := &stubUserGroupRepo{}
	userRepo := &stubUserRepo{}

	dbConn, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	baseRepos := repository.NewRepositories(dbConn)
	baseRepos.ConfigFile = cfRepo
	baseRepos.Resource = resRepo
	baseRepos.Audit = auditRepo
	baseRepos.Project = projectRepo
	baseRepos.UserGroup = userGroupRepo
	baseRepos.User = userRepo

	svc := application.NewConfigFileService(baseRepos)

	k8s.InitTestCluster()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("POST", "/", nil)
	c.Request = req
	c.Set("claims", &types.Claims{Username: "testuser", UserID: "1"})

	utils.SplitYAMLDocuments = func(yamlStr string) []string { return []string{yamlStr} }
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}

	_ = datatypes.JSON([]byte("{}"))

	return svc, cfRepo, resRepo, auditRepo, userRepo, projectRepo, userGroupRepo, c
}
