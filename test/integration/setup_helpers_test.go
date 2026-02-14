//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/linskybing/platform-go/internal/config/db"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/image"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

func ensureDefaultTestEnv() {
	if os.Getenv("JWT_SECRET") == "" {
		_ = os.Setenv("JWT_SECRET", "test-secret")
	}
	if os.Getenv("Issuer") == "" {
		_ = os.Setenv("Issuer", "platform-test")
	}
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
		Username:     username,
		Email:        email,
		Status:       "online",
		IsSuperAdmin: username == "admin",
	}
	if err := db.DB.Create(&u).Error; err != nil {
		panic(fmt.Sprintf("failed to create user %s: %v", username, err))
	}
	return &u
}

func getOrCreateGroup(groupName string) *group.Group {
	var g group.Group
	err := db.DB.Where("name = ?", groupName).First(&g).Error
	if err == nil {
		return &g
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query group %s: %v", groupName, err))
	}

	g = group.Group{
		Name:        groupName,
		Description: fmt.Sprintf("Test group %s", groupName),
	}
	if err := db.DB.Create(&g).Error; err != nil {
		panic(fmt.Sprintf("failed to create group %s: %v", groupName, err))
	}
	return &g
}

func ensureUserGroup(uid string, gid string, role string) {
	var ug group.UserGroup
	err := db.DB.Where("user_id = ? AND group_id = ?", uid, gid).First(&ug).Error
	if err == nil {
		return
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query user_group (uid=%s, gid=%s): %v", uid, gid, err))
	}

	ug = group.UserGroup{
		UserID:  uid,
		GroupID: gid,
		Role:    role,
	}
	if err := db.DB.Create(&ug).Error; err != nil {
		panic(fmt.Sprintf("failed to create user_group (uid=%s, gid=%s, role=%s): %v", uid, gid, role, err))
	}
}

func getOrCreateProject(name string, gid string) *project.Project {
	var p project.Project
	err := db.DB.Preload("ResourcePlan").Where("project_name = ? AND owner_id = ?", name, gid).First(&p).Error
	if err == nil {
		return &p
	}
	if err != gorm.ErrRecordNotFound {
		panic(fmt.Sprintf("failed to query project %s (gid=%s): %v", name, gid, err))
	}

	p = project.Project{
		Name:        name,
		OwnerID:     &gid,
		Description: fmt.Sprintf("Test project %s", name),
	}
	if err := db.DB.Create(&p).Error; err != nil {
		panic(fmt.Sprintf("failed to create project %s (gid=%s): %v", name, gid, err))
	}

	// Create default resource plan
	plan := &project.ResourcePlan{
		ProjectID:  p.ID,
		WeekWindow: "[0,604800)",
		GPULimit:   10,
	}
	db.DB.Create(plan)

	// Reload to get generated fields like Path and preloaded ResourcePlan
	db.DB.Preload("ResourcePlan").First(&p, "p_id = ?", p.ID)
	return &p
}

func allowImageGlobally(name, tag string) {
	parts := strings.Split(name, "/")
	var namespace, repoName string
	if len(parts) >= 2 {
		namespace = parts[0]
		repoName = strings.Join(parts[1:], "/")
	} else {
		namespace = "library"
		repoName = name
	}

	repo := image.ContainerRepository{
		Namespace: namespace,
		Name:      repoName,
		FullName:  name,
	}
	db.DB.Where(image.ContainerRepository{FullName: name}).FirstOrCreate(&repo)

	tagEntity := image.ContainerTag{
		RepositoryID: repo.ID,
		Name:         tag,
	}
	db.DB.Where(image.ContainerTag{RepositoryID: repo.ID, Name: tag}).FirstOrCreate(&tagEntity)

	rule := image.ImageAllowList{
		RepositoryID: repo.ID,
		TagID:        &tagEntity.ID,
		IsEnabled:    true,
		CreatedBy:    getOrCreateUser("admin", "admin@test.com").ID,
	}
	db.DB.Where(image.ImageAllowList{RepositoryID: repo.ID, TagID: &tagEntity.ID}).FirstOrCreate(&rule)
}
