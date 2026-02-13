package plugin

import (
"context"
"errors"
"testing"

"github.com/gin-gonic/gin"
"github.com/linskybing/platform-go/internal/events"
"github.com/linskybing/platform-go/pkg/cache"
"github.com/stretchr/testify/assert"
"gorm.io/driver/sqlite"
"gorm.io/gorm"
)

// MockPlugin implements Plugin interface for testing
type MockPlugin struct {
Initialized   bool
Registered    bool
ShutdownCalled bool
ShouldFailInit bool
ShouldFailShutdown bool
}

func (m *MockPlugin) Name() string {
return "mock-plugin"
}

func (m *MockPlugin) Version() string {
return "1.0.0"
}

func (m *MockPlugin) Init(ctx PluginContext) error {
if m.ShouldFailInit {
return errors.New("init failed")
}
m.Initialized = true
return nil
}

func (m *MockPlugin) RegisterRoutes(r *gin.RouterGroup) {
m.Registered = true
// Test adding a route
r.GET("/ping", func(c *gin.Context) {
c.JSON(200, gin.H{"message": "pong"})
})
}

func (m *MockPlugin) RegisterMigrations() []Migration {
return []Migration{
{ID: "test-migration", Up: []string{"CREATE TABLE x (id TEXT)"}},
}
}

func (m *MockPlugin) RegisterEvents(bus events.EventBus) {
}

func (m *MockPlugin) Shutdown() error {
if m.ShouldFailShutdown {
return errors.New("shutdown failed")
}
m.ShutdownCalled = true
return nil
}

func setupTestDB() *gorm.DB {
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
return db
}

func TestPluginRegistry(t *testing.T) {
// Reset registry
GlobalRegistry = &Registry{
plugins: make(map[string]Plugin),
}

mockPlugin := &MockPlugin{}
Register(mockPlugin)

// Test duplicate registration panic
assert.Panics(t, func() {
Register(mockPlugin)
})

// Verify registration
routes := GlobalRegistry.GetRoutes()
assert.Contains(t, routes, mockPlugin)

// Test InitAll
db := setupTestDB()
ctx := PluginContext{DB: db}

err := GlobalRegistry.InitAll(ctx)
assert.NoError(t, err)
assert.True(t, mockPlugin.Initialized)

// Test InitAll Error
// We can't register with same name, so override name via method not possible on struct pointer without embedding or modifying struct.
// But Register uses .Name(), which returns "mock-plugin". 
// Let's make a new struct type for failing one or make name dynamic.

// Reset registry for error test
GlobalRegistry = &Registry{
plugins: make(map[string]Plugin),
}
Register(&MockPlugin{ShouldFailInit: true})
err = GlobalRegistry.InitAll(ctx)
assert.Error(t, err)
assert.Contains(t, err.Error(), "init failed")

// Test Shutdown
GlobalRegistry = &Registry{
plugins: make(map[string]Plugin),
}
mockPlugin = &MockPlugin{}
Register(mockPlugin)
GlobalRegistry.ShutdownAll()
assert.True(t, mockPlugin.ShutdownCalled)

// Test Shutdown Error (check logs? no return error from ShutdownAll)
GlobalRegistry = &Registry{
plugins: make(map[string]Plugin),
}
Register(&MockPlugin{ShouldFailShutdown: true})
GlobalRegistry.ShutdownAll() // Should not panic
}

func TestManager(t *testing.T) {
// Setup
GlobalRegistry = &Registry{
plugins: make(map[string]Plugin),
}
mockPlugin := &MockPlugin{}
Register(mockPlugin)

db := setupTestDB()
cacheSvc := cache.NewNoop()

manager := NewManager(db, cacheSvc)

// Test Init
err := manager.Init()
assert.NoError(t, err)
assert.True(t, mockPlugin.Initialized)

// Test RegisterRoutes
gin.SetMode(gin.TestMode)
r := gin.New()
api := r.Group("/api")
manager.RegisterRoutes(api)
assert.True(t, mockPlugin.Registered)

// Test Shutdown
manager.Shutdown()
assert.True(t, mockPlugin.ShutdownCalled)
}

func TestHookRegistry(t *testing.T) {
registry := NewHookRegistry()
executed := false

// Register hook
registry.Register(HookBeforeCreate, func(ctx context.Context, resourceType string, data interface{}) error {
executed = true
if resourceType == "error" {
return errors.New("hook error")
}
return nil
})

// Execute success
err := registry.Execute(context.Background(), HookBeforeCreate, "job", nil)
assert.NoError(t, err)
assert.True(t, executed)

// Execute error
err = registry.Execute(context.Background(), HookBeforeCreate, "error", nil)
assert.Error(t, err)

// Execute no-op (no hooks for type)
err = registry.Execute(context.Background(), HookAfterCreate, "job", nil)
assert.NoError(t, err)
}
