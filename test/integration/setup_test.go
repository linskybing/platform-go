//go:build integration
// +build integration

package integration

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
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
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
	"gorm.io/gorm"
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

	// Enable K8S mock mode for integration tests
	if os.Getenv("K8S_MOCK") == "true" {
		k8s.SetMockMode(true)
	}

	config.LoadConfig()
	config.InitK8sConfig()

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

	routes.RegisterRoutes(router, db.DB, cache.NewNoop())

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
		&configfile.ConfigFile{},
		&resource.Resource{},
		&form.Form{},
		&form.FormMessage{},
		&audit.AuditLog{},
		&image.ContainerRepository{},
		&image.ContainerTag{},
		&image.ImageAllowList{},
		&image.ImageRequest{},
		&image.ClusterImageStatus{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	db.CreateViews()
}

func ensureDefaultTestEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("Issuer") == "" {
		_ = os.Setenv("Issuer", "platform-test")
	}
	// Enable K8S mock by default for integration tests
	if os.Getenv("K8S_MOCK") == "" {
		_ = os.Setenv("K8S_MOCK", "true")
	}
	if os.Getenv("SKIP_K8S_TESTS") == "" {
		_ = os.Setenv("SKIP_K8S_TESTS", "true")
	}
}

func seedEnvFromDatabaseURL() {
	if os.Getenv("DB_HOST") != "" && os.Getenv("DB_NAME") != "" {
		return
	}
	if os.Getenv("DATABASE_URL") == "" {
		return
	}

	u, err := url.Parse(os.Getenv("DATABASE_URL"))
	if err != nil {
		return
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}

	userName := ""
	password := ""
	if u.User != nil {
		userName = u.User.Username()
		password, _ = u.User.Password()
	}

	dbName := strings.TrimPrefix(u.Path, "/")

	if os.Getenv("DB_HOST") == "" {
		_ = os.Setenv("DB_HOST", host)
	}
	if os.Getenv("DB_PORT") == "" {
		_ = os.Setenv("DB_PORT", port)
	}
	if os.Getenv("DB_USER") == "" {
		_ = os.Setenv("DB_USER", userName)
	}
	if os.Getenv("DB_PASSWORD") == "" {
		_ = os.Setenv("DB_PASSWORD", password)
	}
	if os.Getenv("DB_NAME") == "" {
		_ = os.Setenv("DB_NAME", dbName)
	}
}

func getOrCreateUser(username, email string) *user.User {
	var u user.User
	err := db.DB.Where("username = ?", username).First(&u).Error
	if err == nil {
		return &u
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query user %s: %v", username, err))
	}

	u = user.User{
		Username: username,
		Email:    &email,
		Status:   "online",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		panic(fmt.Sprintf("failed to create user %s: %v", username, err))
	}
	return &u
}

func getOrCreateGroup(groupName string) *group.Group {
	var g group.Group
	err := db.DB.Where("group_name = ?", groupName).First(&g).Error
	if err == nil {
		return &g
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query group %s: %v", groupName, err))
	}

	g = group.Group{
		GroupName:   groupName,
		Description: fmt.Sprintf("Test group %s", groupName),
	}
	if err := db.DB.Create(&g).Error; err != nil {
		panic(fmt.Sprintf("failed to create group %s: %v", groupName, err))
	}
	return &g
}

func ensureUserGroup(uid string, gid string, role string) {
	var ug group.UserGroup
	err := db.DB.Where("u_id = ? AND g_id = ?", uid, gid).First(&ug).Error
	if err == nil {
		return
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query user_group (uid=%s, gid=%s): %v", uid, gid, err))
	}

	ug = group.UserGroup{
		UID:  uid,
		GID:  gid,
		Role: role,
	}
	if err := db.DB.Create(&ug).Error; err != nil {
		panic(fmt.Sprintf("failed to create user_group (uid=%s, gid=%s, role=%s): %v", uid, gid, role, err))
	}
}

func getOrCreateProject(name string, gid string) *project.Project {
	var p project.Project
	err := db.DB.Where("project_name = ? AND g_id = ?", name, gid).First(&p).Error
	if err == nil {
		return &p
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query project %s (gid=%s): %v", name, gid, err))
	}

	p = project.Project{
		ProjectName: name,
		GID:         gid,
		Description: fmt.Sprintf("Test project %s", name),
	}
	if err := db.DB.Create(&p).Error; err != nil {
		panic(fmt.Sprintf("failed to create project %s (gid=%s): %v", name, gid, err))
	}
	return &p
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().Unix(), rand.Intn(10000))
}

func randomEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.local", prefix, rand.Intn(10000))
}
