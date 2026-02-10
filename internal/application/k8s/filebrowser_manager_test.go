package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/pkg/filebrowser"
)

// mockSessionManager implements filebrowser.SessionManager for tests
type mockSessionManager struct {
	stopped []struct{ namespace, pod, svc string }
	stopErr error
}

func (m *mockSessionManager) GetOrCreate(ctx context.Context, cfg *filebrowser.Config) (string, error) {
	return "30000", nil
}
func (m *mockSessionManager) Start(ctx context.Context, cfg *filebrowser.Config) (string, error) {
	return "30000", nil
}
func (m *mockSessionManager) Stop(ctx context.Context, namespace, podName, serviceName string) error {
	m.stopped = append(m.stopped, struct{ namespace, pod, svc string }{namespace, podName, serviceName})
	return m.stopErr
}

// mockPerm provides fake permission responses
type mockPerm struct {
	perm *storage.GroupStoragePermission
	err  error
}

func (m *mockPerm) GetUserPermission(ctx context.Context, userID, groupID, pvcID string) (*storage.GroupStoragePermission, error) {
	return m.perm, m.err
}

func TestStopFileBrowser_NoPermission(t *testing.T) {
	pm := &mockPerm{perm: &storage.GroupStoragePermission{Permission: storage.PermissionNone}}
	fbm := NewFileBrowserManager(pm)
	// replace session manager with mock
	ms := &mockSessionManager{}
	fbm.sessionMgr = ms

	// use k8sclient nil path: listPVCsByID will return mock-pvc
	err := fbm.StopFileBrowser(context.Background(), "g1", "group-g1-abc", "user1")
	if err == nil {
		t.Fatal("expected error when user has no permission")
	}
}

func TestStopFileBrowser_StopRW(t *testing.T) {
	pm := &mockPerm{perm: &storage.GroupStoragePermission{Permission: storage.PermissionReadWrite}}
	fbm := NewFileBrowserManager(pm)
	ms := &mockSessionManager{}
	fbm.sessionMgr = ms

	err := fbm.StopFileBrowser(context.Background(), "g1", "group-g1-abc", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ms.stopped) == 0 {
		t.Fatalf("expected Stop to be called on session manager")
	}
}

func TestStopFileBrowser_StopFailure(t *testing.T) {
	pm := &mockPerm{perm: &storage.GroupStoragePermission{Permission: storage.PermissionReadOnly}}
	fbm := NewFileBrowserManager(pm)
	ms := &mockSessionManager{stopErr: errors.New("k8s delete failed")}
	fbm.sessionMgr = ms

	err := fbm.StopFileBrowser(context.Background(), "g1", "group-g1-abc", "user1")
	if err == nil {
		t.Fatal("expected error when session manager fails to stop")
	}
}
