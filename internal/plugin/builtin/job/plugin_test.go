package job

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/plugin"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	return db
}

func TestJobPluginLifecycle(t *testing.T) {
	// Setup
	p := NewJobPlugin()

	// Test Name & Version
	assert.Equal(t, "job-manager", p.Name())
	assert.Equal(t, "1.0.0", p.Version())

	// Test Init
	db := setupTestDB()
	ctx := plugin.PluginContext{
		DB:    db,
		Cache: cache.NewNoop(),
	}

	err := p.Init(ctx)
	assert.NoError(t, err)

	// Test Route Registration
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rg := r.Group("/test")

	// Should not panic
	p.RegisterRoutes(rg)

	// Test other methods
	assert.Nil(t, p.RegisterMigrations())
	p.RegisterEvents(nil)
	assert.NoError(t, p.Shutdown())
}

func TestRefactorVerification(t *testing.T) {
	// This test confirms that the plugin correctly implements the interface
	var _ plugin.Plugin = (*JobPlugin)(nil)
}
