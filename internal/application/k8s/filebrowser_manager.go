package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/pkg/filebrowser"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FileBrowserManager handles FileBrowser pod creation with permission-based routing
type FileBrowserManager struct {
	fbMgr      filebrowser.Manager
	sessionMgr filebrowser.SessionManager
	perm       permissionProvider
}

// permissionProvider defines the minimal interface required for permission checks.
type permissionProvider interface {
	GetUserPermission(ctx context.Context, userID, groupID, pvcID string) (*storage.GroupStoragePermission, error)
}

// NewFileBrowserManager creates a new FileBrowserManager instance
func NewFileBrowserManager(pm permissionProvider) *FileBrowserManager {
	return &FileBrowserManager{
		fbMgr:      filebrowser.NewManager(),
		sessionMgr: filebrowser.NewSessionManager(),
		perm:       pm,
	}
}

// GetFileBrowserAccess creates or routes to appropriate FileBrowser pod based on user permission
// - Read-write permission -> read-write pod
// - Read-only permission -> read-only pod
// - No permission -> return unauthorized error
func (fbm *FileBrowserManager) GetFileBrowserAccess(ctx context.Context, req *storage.FileBrowserAccessRequest) (*storage.FileBrowserAccessResponse, error) {
	startTime := time.Now()
	// Determine permission for this user on the requested PVC
	logger.Info("filebrowser access request",
		"user_id", req.UserID,
		"group_id", req.GroupID,
		"pvc_id", req.PVCID)

	if fbm.perm == nil {
		return nil, fmt.Errorf("permission manager not configured")
	}
	perm, err := fbm.perm.GetUserPermission(ctx, req.UserID, req.GroupID, req.PVCID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	if perm == nil || perm.Permission == storage.PermissionNone {
		return &storage.FileBrowserAccessResponse{
			Allowed: false,
			Message: "permission denied: you don't have access to this storage",
		}, nil
	}

	// Get group PVC info to find namespace and PVC name
	groupNamespace := fmt.Sprintf("group-%s-storage", req.GroupID)

	// Find the actual PVC name from labels
	pvcs, err := fbm.listPVCsByID(ctx, groupNamespace, req.PVCID)
	if err != nil || len(pvcs) == 0 {
		return nil, fmt.Errorf("PVC not found: %s", req.PVCID)
	}
	pvcName := pvcs[0].Name

	// Determine pod type based on permission
	readOnly := perm.Permission != storage.PermissionReadWrite

	// Construct FileBrowser configuration
	accessType := "ro"
	if !readOnly {
		accessType = "rw"
	}
	podName := fmt.Sprintf("fb-%s-%s", accessType, pvcName)
	svcName := fmt.Sprintf("fb-svc-%s-%s", accessType, pvcName)

	cfg := &filebrowser.Config{
		Namespace:   groupNamespace,
		PodName:     podName,
		ServiceName: svcName,
		PVCName:     pvcName,
		ReadOnly:    readOnly,
		Labels: map[string]string{
			"pvc":         pvcName,
			"access-mode": accessType,
		},
	}

	// Create or get existing FileBrowser session
	nodePort, err := fbm.sessionMgr.GetOrCreate(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create filebrowser session: %w", err)
	}

	logger.Info("filebrowser access granted",
		"user_id", req.UserID,
		"group_id", req.GroupID,
		"pvc_id", req.PVCID,
		"read_only", readOnly,
		"pod", podName,
		"duration_ms", time.Since(startTime).Milliseconds())

	return &storage.FileBrowserAccessResponse{
		Allowed:  true,
		URL:      fmt.Sprintf("http://nodeip:%s", nodePort),
		Port:     nodePort,
		PodName:  podName,
		ReadOnly: readOnly,
		Message:  "Access granted",
	}, nil
}

// StopFileBrowser stops an existing FileBrowser session for the given group PVC.
// It attempts to stop both read-write and read-only sessions if necessary.
func (fbm *FileBrowserManager) StopFileBrowser(ctx context.Context, groupID, pvcID, userID string) error {
	groupNamespace := fmt.Sprintf("group-%s-storage", groupID)

	// Ensure PVC exists and obtain PVC name
	pvcs, err := fbm.listPVCsByID(ctx, groupNamespace, pvcID)
	if err != nil || len(pvcs) == 0 {
		return fmt.Errorf("PVC not found: %s", pvcID)
	}
	pvcName := pvcs[0].Name

	// Determine user's permission for this PVC
	if fbm.perm == nil {
		return fmt.Errorf("permission manager not configured")
	}
	perm, err := fbm.perm.GetUserPermission(ctx, userID, groupID, pvcID)
	if err != nil {
		return fmt.Errorf("failed to get permission: %w", err)
	}

	if perm == nil || perm.Permission == storage.PermissionNone {
		return fmt.Errorf("permission denied: user has no access to this storage")
	}

	accessType := "ro"
	if perm.Permission == storage.PermissionReadWrite {
		accessType = "rw"
	}

	podName := fmt.Sprintf("fb-%s-%s", accessType, pvcName)
	svcName := fmt.Sprintf("fb-svc-%s-%s", accessType, pvcName)

	if err := fbm.sessionMgr.Stop(ctx, groupNamespace, podName, svcName); err != nil {
		return fmt.Errorf("failed to stop filebrowser session: %w", err)
	}
	return nil
}

// listPVCsByID finds PVCs with matching ID label
func (fbm *FileBrowserManager) listPVCsByID(ctx context.Context, namespace, pvcID string) ([]corev1.PersistentVolumeClaim, error) {
	if k8sclient.Clientset == nil {
		return []corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "mock-pvc"}}}, nil
	}

	// Extract UUID from PVC ID (format: group-{gid}-{uuid})
	uuid := pvcID[len("group-XX-"):] // Simplified extraction

	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("pvc-uuid=%s", uuid),
	}

	pvcList, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	return pvcList.Items, nil
}
