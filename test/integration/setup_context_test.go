//go:build integration
// +build integration

package integration

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/api/routes"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/form"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/plugin"
	jobplugin "github.com/linskybing/platform-go/internal/plugin/builtin/job"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
)

type TestContext struct {
	Router        *gin.Engine
	AdminToken    string
	ManagerToken  string
	UserToken     string
	TestAdmin     *user.User
	TestManager   *user.User
	TestUser      *user.User
	TestGroup     *group.Group
	SuperGroup    *group.Group
	TestProject   *project.Project
	TestNamespace string
}

var (
	testContext *TestContext
	setupOnce   sync.Once
)

func init() {
	ensureDefaultTestEnv()
}

func GetTestContext() *TestContext {
	setupOnce.Do(func() {
		testContext = setupTestContext()
	})
	return testContext
}

func setupTestContext() *TestContext {
	seedEnvFromDatabaseURL()
	ensureDefaultTestEnv()

	if os.Getenv("K8S_MOCK") == "true" {
		k8s.SetMockMode(true)
	}

	config.LoadConfig()

	middleware.Init()
	initDatabase()

	repos := repository.NewRepositories(db.DB)

	adminUser := getOrCreateUser("admin", "admin@test.com")
	superGroup := getOrCreateGroup("super")
	ensureUserGroup(adminUser.UID, superGroup.GID, "admin")

	managerUser := getOrCreateUser(randomName("manager"), randomEmail("manager"))
	regularUser := getOrCreateUser(randomName("user"), randomEmail("user"))

	testGroup := getOrCreateGroup(randomName("group"))
	ensureUserGroup(managerUser.UID, testGroup.GID, "manager")
	ensureUserGroup(regularUser.UID, testGroup.GID, "user")

	testProject := getOrCreateProject(randomName("project"), testGroup.GID)

	adminToken, _, err := middleware.GenerateToken(adminUser.UID, adminUser.Username, time.Hour, repos.UserGroup)
	if err != nil {
		panic(fmt.Sprintf("failed to generate admin token: %v", err))
	}

	managerToken, _, err := middleware.GenerateToken(managerUser.UID, managerUser.Username, time.Hour, repos.UserGroup)
	if err != nil {
		panic(fmt.Sprintf("failed to generate manager token: %v", err))
	}

	userToken, _, err := middleware.GenerateToken(regularUser.UID, regularUser.Username, time.Hour, repos.UserGroup)
	if err != nil {
		panic(fmt.Sprintf("failed to generate user token: %v", err))
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())

	cacheSvc := cache.NewNoop()
	routes.RegisterRoutes(router, db.DB, cacheSvc)

	jobP := jobplugin.NewJobPlugin()
	func() {
		defer func() { _ = recover() }()
		plugin.Register(jobP)
	}()

	mgr := plugin.NewManager(db.DB, cacheSvc)
	if err := mgr.Init(); err != nil {
		panic(fmt.Sprintf("failed to init plugins: %v", err))
	}

	authGroup := router.Group("/")
	authGroup.Use(middleware.JWTAuthMiddleware())
	mgr.RegisterRoutes(authGroup)

	return &TestContext{
		Router:        router,
		AdminToken:    adminToken,
		ManagerToken:  managerToken,
		UserToken:     userToken,
		TestAdmin:     adminUser,
		TestManager:   managerUser,
		TestUser:      regularUser,
		TestGroup:     testGroup,
		SuperGroup:    superGroup,
		TestProject:   testProject,
		TestNamespace: fmt.Sprintf("test-%d", time.Now().UnixNano()),
	}
}

func initDatabase() {
	db.Init()

	if err := db.DB.AutoMigrate(
		&user.User{},
		&group.Group{},
		&group.UserGroup{},
		&project.Project{},
		&configfile.ConfigBlob{},
		&configfile.ConfigCommit{},
		&resource.Resource{},
		&form.Form{},
		&form.FormMessage{},
		&audit.AuditLog{},
		&image.ContainerRepository{},
		&image.ContainerTag{},
		&image.ImageAllowList{},
		&image.ImageRequest{},
		&image.ClusterImageStatus{},
		&job.Job{},
		&storage.Storage{},
		&storage.GroupStoragePermission{},
		&storage.GroupStorageAccessPolicy{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().Unix(), rand.Intn(10000))
}

func randomEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.local", prefix, rand.Intn(10000))
}
