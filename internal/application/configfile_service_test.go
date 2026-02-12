package application_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/domain/view"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/internal/repository/mock"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMocks(t *testing.T) (*application.ConfigFileService, *mock.MockConfigFileRepo,
	*mock.MockResourceRepo, *mock.MockAuditRepo,
	*mock.MockUserRepo, *mock.MockProjectRepo, *mock.MockUserGroupRepo, *gin.Context) {

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockCF := mock.NewMockConfigFileRepo(ctrl)
	mockRes := mock.NewMockResourceRepo(ctrl)
	mockAudit := mock.NewMockAuditRepo(ctrl)
	mockProject := mock.NewMockProjectRepo(ctrl)
	mockUserGroup := mock.NewMockUserGroupRepo(ctrl)
	mockUser := mock.NewMockUserRepo(ctrl)
	// create base repos with an in-memory sqlite gorm DB so Begin() is safe, then inject mocks
	dbConn, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	baseRepos := repository.NewRepositories(dbConn)
	baseRepos.ConfigFile = mockCF
	baseRepos.Resource = mockRes
	baseRepos.Audit = mockAudit
	baseRepos.Project = mockProject
	baseRepos.UserGroup = mockUserGroup
	baseRepos.User = mockUser
	repos := baseRepos
	// (db already set in baseRepos)
	svc := application.NewConfigFileService(repos)

	// initialize fake k8s client for functions that require Clientset
	k8s.InitTestCluster()

	// gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("POST", "/", nil)
	c.Request = req
	c.Set("claims", &types.Claims{Username: "testuser", UserID: "1"})

	// mock utils (k8s functions use mock behavior when Mapper/DynamicClient/Clientset are nil)
	utils.SplitYAMLDocuments = func(yamlStr string) []string { return []string{yamlStr} }
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}

	// Ensure WithTx returns the same mock so transactional calls use the expected mock methods
	mockCF.EXPECT().WithTx(gomock.Any()).DoAndReturn(func(tx *gorm.DB) repository.ConfigFileRepo {
		return mockCF
	}).AnyTimes()
	mockRes.EXPECT().WithTx(gomock.Any()).DoAndReturn(func(tx *gorm.DB) repository.ResourceRepo {
		fmt.Println("[MOCK] ResourceRepo.WithTx called")
		return mockRes
	}).AnyTimes()
	// CreateConfigFile may or may not be invoked depending on internal transaction flow in service; tests set expectations where needed.

	return svc, mockCF, mockRes, mockAudit, mockUser, mockProject, mockUserGroup, c
}

func TestCreateConfigFile_Success(t *testing.T) {
	svc, mockCF, mockRes, mockAudit, _, _, _, c := setupMocks(t)

	mockCF.EXPECT().CreateConfigFile(gomock.Any()).Return(nil)
	mockRes.EXPECT().CreateResource(gomock.Any()).Return(nil).AnyTimes()
	mockAudit.EXPECT().CreateAuditLog(gomock.Any()).Return(nil).AnyTimes()

	input := configfile.CreateConfigFileInput{
		Filename:  "test.yaml",
		RawYaml:   "apiVersion: v1\nkind: Pod\nmetadata:\n  name: testpod",
		ProjectID: "1",
	}

	cf, err := svc.CreateConfigFile(c.Request.Context(), input, c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.Filename != "test.yaml" {
		t.Fatalf("expected filename test.yaml, got %s", cf.Filename)
	}
}

func TestCreateConfigFile_NoYAMLDocuments(t *testing.T) {
	svc, _, _, _, _, _, _, c := setupMocks(t)

	utils.SplitYAMLDocuments = func(yamlStr string) []string { return []string{} }

	input := configfile.CreateConfigFileInput{
		Filename:  "test.yaml",
		RawYaml:   "",
		ProjectID: "1",
	}

	_, err := svc.CreateConfigFile(c.Request.Context(), input, c.MustGet("claims").(*types.Claims))
	if !errors.Is(err, application.ErrNoValidYAMLDocument) {
		t.Fatalf("expected ErrNoValidYAMLDocument, got %v", err)
	}
}

func TestUpdateConfigFile_Success(t *testing.T) {
	svc, mockCF, mockRes, mockAudit, _, _, _, c := setupMocks(t)

	// Mock original ConfigFile
	existingCF := &configfile.ConfigFile{
		CFID:      "1",
		ProjectID: "1",
		Filename:  "old.yaml",
	}
	mockCF.EXPECT().GetConfigFileByID("1").Return(existingCF, nil)
	mockCF.EXPECT().UpdateConfigFile(gomock.Any()).Return(nil)

	// Mock Resource
	mockRes.EXPECT().ListResourcesByConfigFileID("1").Return([]resource.Resource{}, nil)
	mockRes.EXPECT().CreateResource(gomock.Any()).Return(nil).AnyTimes()
	mockRes.EXPECT().UpdateResource(gomock.Any()).Return(nil).AnyTimes()
	mockRes.EXPECT().DeleteResource(gomock.Any()).Return(nil).AnyTimes()

	// Mock User repo listing
	// no user list required for update path in current implementation

	// Mock Audit
	mockAudit.EXPECT().CreateAuditLog(gomock.Any()).Return(nil).AnyTimes()

	// Mock utils: keep split behavior consistent so actual YAML is processed
	utils.SplitYAMLDocuments = func(yamlStr string) []string {
		return []string{yamlStr}
	}
	// use actual k8s.ValidateK8sJSON implementation (pure function)
	utils.LogAuditWithConsole = func(c *gin.Context, action, resourceType, resourceID string, oldData, newData interface{}, msg string, repos repository.AuditRepo) {
	}
	// k8s.DeleteByJson will use mock behavior when k8s clients are nil

	filename := "new.yaml"
	rawYaml := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: testpod"
	input := configfile.ConfigFileUpdateDTO{
		Filename: &filename,
		RawYaml:  &rawYaml,
	}

	cf, err := svc.UpdateConfigFile(c.Request.Context(), "1", input, c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cf.Filename != "new.yaml" {
		t.Fatalf("expected filename new.yaml, got %s", cf.Filename)
	}
}

func TestDeleteConfigFile_Success(t *testing.T) {
	svc, mockCF, mockRes, mockAudit, mockUser, _, _, c := setupMocks(t)

	mockCF.EXPECT().GetConfigFileByID("1").Return(&configfile.ConfigFile{
		CFID: "1", ProjectID: "1", Filename: "test.yaml",
	}, nil).AnyTimes()

	mockRes.EXPECT().ListResourcesByConfigFileID("1").Return([]resource.Resource{
		{RID: "10", Name: "res1"},
	}, nil).AnyTimes()

	mockUser.EXPECT().ListUsersByProjectID("1").Return([]view.ProjectUserView{
		{Username: "user1"},
	}, nil)

	mockRes.EXPECT().DeleteResource("10").Return(nil).AnyTimes()
	mockCF.EXPECT().DeleteConfigFile("1").Return(nil).AnyTimes()
	mockAudit.EXPECT().CreateAuditLog(gomock.Any()).Return(nil).AnyTimes()

	// k8s.DeleteByJson is a function that uses mock behavior when k8s clients are nil, so no override needed

	err := svc.DeleteConfigFile(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateInstance_Success(t *testing.T) {
	svc, mockCF, mockRes, _, _, mockProject, mockUserGroup, c := setupMocks(t)

	mockRes.EXPECT().ListResourcesByConfigFileID("1").Return([]resource.Resource{{RID: "1", ParsedYAML: datatypes.JSON([]byte("{}"))}}, nil)
	mockCF.EXPECT().GetConfigFileByID("1").Return(&configfile.ConfigFile{CFID: "1", ProjectID: "1"}, nil)
	mockProject.EXPECT().GetProjectByID("1").Return(project.Project{PID: "1", GID: "10"}, nil).AnyTimes()
	mockUserGroup.EXPECT().GetUserGroup("1", "10").Return(group.UserGroup{UID: "1", GID: "10", Role: "admin"}, nil).AnyTimes()

	err := svc.CreateInstance(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteInstance_Success(t *testing.T) {
	svc, mockCF, mockRes, _, _, _, _, c := setupMocks(t)

	mockRes.EXPECT().ListResourcesByConfigFileID("1").Return([]resource.Resource{{RID: "1", ParsedYAML: datatypes.JSON([]byte("{}"))}}, nil)
	mockCF.EXPECT().GetConfigFileByID("1").Return(&configfile.ConfigFile{CFID: "1", ProjectID: "1"}, nil)

	err := svc.DeleteInstance(c.Request.Context(), "1", c.MustGet("claims").(*types.Claims))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteConfigFileInstance_Success(t *testing.T) {
	svc, mockCF, mockRes, _, mockUser, _, _, _ := setupMocks(t)

	// Mock ConfigFile
	mockCF.EXPECT().GetConfigFileByID("1").Return(&configfile.ConfigFile{
		CFID:      "1",
		ProjectID: "1",
		Filename:  "test.yaml",
	}, nil)

	// Mock Resource
	mockRes.EXPECT().ListResourcesByConfigFileID("1").Return([]resource.Resource{
		{RID: "1", Name: "res1"},
	}, nil)

	// Mock User repo listing
	mockUser.EXPECT().ListUsersByProjectID("1").Return([]view.ProjectUserView{
		{Username: "user1"},
	}, nil)

	// k8s.FormatNamespaceName and k8s.DeleteByJson use deterministic behavior / mock when clients are nil

	err := svc.DeleteConfigFileInstance("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigFileRead(t *testing.T) {
	svc, mockCF, _, _, _, _, _, _ := setupMocks(t)

	t.Run("ListConfigFiles", func(t *testing.T) {
		cfs := []configfile.ConfigFile{{CFID: "1", Filename: "f1"}}
		mockCF.EXPECT().ListConfigFiles().Return(cfs, nil)

		res, err := svc.ListConfigFiles()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 config file, got %d", len(res))
		}
	})

	t.Run("GetConfigFile", func(t *testing.T) {
		cf := &configfile.ConfigFile{CFID: "1", Filename: "f1"}
		mockCF.EXPECT().GetConfigFileByID("1").Return(cf, nil)

		res, err := svc.GetConfigFile("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Filename != "f1" {
			t.Fatalf("expected f1, got %s", res.Filename)
		}
	})

	t.Run("ListConfigFilesByProjectID", func(t *testing.T) {
		cfs := []configfile.ConfigFile{{CFID: "1", Filename: "f1"}}
		mockCF.EXPECT().GetConfigFilesByProjectID("10").Return(cfs, nil)

		res, err := svc.ListConfigFilesByProjectID("10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 config file, got %d", len(res))
		}
	})
}
