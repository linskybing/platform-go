package configfile

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
)

// --- Helpers for Deployment ---

func (s *ConfigFileService) prepareNamespaceAndProject(ctx context.Context, cf *configfile.ConfigFile, claims *types.Claims) (string, project.Project, error) {
	safeUsername := k8s.ToSafeK8sName(claims.Username)
	targetNs := k8s.FormatNamespaceName(cf.ProjectID, safeUsername)

	if err := k8s.EnsureNamespaceExists(targetNs); err != nil {
		return "", project.Project{}, fmt.Errorf("failed to ensure namespace %s: %w", targetNs, err)
	}

	p, err := s.Repos.Project.GetProjectByID(cf.ProjectID)
	if err != nil {
		return "", project.Project{}, err
	}
	return targetNs, p, nil
}

func (s *ConfigFileService) bindProjectAndUserVolumes(targetNs string, project project.Project, claims *types.Claims) (string, string) {
	safeUsername := k8s.ToSafeK8sName(claims.Username)
	userStorageNs := fmt.Sprintf(config.UserStorageNs, safeUsername)
	userPvcName := fmt.Sprintf(config.UserStoragePVC, safeUsername)
	groupStorageNs := k8s.GenerateSafeResourceName("group", project.ProjectName, project.PID)
	groupPvcName := fmt.Sprintf("group-%s-disk", project.PID)

	targetUserPvcName := userPvcName
	if err := k8s.MountExistingVolumeToNamespace(userStorageNs, userPvcName, targetNs, targetUserPvcName); err != nil {
		slog.Warn("failed to bind user volume",
			"username", claims.Username,
			"namespace", userStorageNs,
			"error", err)
	}

	targetGroupPvcName := groupPvcName
	if err := k8s.MountExistingVolumeToNamespace(groupStorageNs, groupPvcName, targetNs, targetGroupPvcName); err != nil {
		slog.Warn("failed to bind group volume",
			"project_id", project.PID,
			"namespace", groupStorageNs,
			"error", err)
	}

	return targetUserPvcName, targetGroupPvcName
}

func (s *ConfigFileService) determineReadOnlyEnforcement(claims *types.Claims, project project.Project) (bool, error) {
	if claims.IsAdmin {
		return false, nil
	}
	ug, err := s.Repos.UserGroup.GetUserGroup(claims.UserID, project.GID)
	if err != nil {
		// If user is not in group, default to safe (Enforce RO) or error?
		// Assuming error means access denied usually, but let's be strict.
		return true, err
	}
	// Only managers and admins get write access
	return ug.Role != "manager" && ug.Role != "admin", nil
}

func (s *ConfigFileService) buildTemplateValues(cf *configfile.ConfigFile, namespace, userPvc, groupPvc string, claims *types.Claims) map[string]string {
	return map[string]string{
		"username":         k8s.ToSafeK8sName(claims.Username),
		"originalUsername": claims.Username,
		"safeUsername":     k8s.ToSafeK8sName(claims.Username),
		"namespace":        namespace,
		"projectId":        cf.ProjectID,
		"userVolume":       userPvc,
		"groupVolume":      groupPvc,
	}
}
