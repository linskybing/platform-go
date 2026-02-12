package configfile

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/resource"
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

func (s *ConfigFileService) bindProjectAndUserVolumes(ctx context.Context, targetNs string, project project.Project, claims *types.Claims, resources []resource.Resource) (string, string, error) {
	claimsMap := collectPVCClaimNames(resources)
	if len(claimsMap) == 0 {
		return "", "", nil
	}

	usesUserVolume := false
	usesGroupVolumeTemplate := false
	groupPVCNames := make(map[string]struct{})

	for claim := range claimsMap {
		trimmed := strings.TrimSpace(claim)
		if strings.Contains(trimmed, "userVolume") {
			usesUserVolume = true
			continue
		}
		if strings.Contains(trimmed, "groupVolume") {
			usesGroupVolumeTemplate = true
			continue
		}
		if trimmed != "" {
			groupPVCNames[trimmed] = struct{}{}
		}
	}

	// Resolve group storages for this project's group
	groupStorages := []struct {
		ID      string
		PVCName string
		GroupID string
	}{}
	if s.Repos != nil && s.Repos.Storage != nil {
		if list, err := s.Repos.Storage.ListGroupStorageByGID(ctx, project.GID); err == nil {
			for _, gs := range list {
				groupStorages = append(groupStorages, struct {
					ID      string
					PVCName string
					GroupID string
				}{ID: gs.ID, PVCName: gs.PVCName, GroupID: gs.GroupID})
			}
		} else if len(groupPVCNames) > 0 || usesGroupVolumeTemplate {
			return "", "", fmt.Errorf("group storage repo not configured")
		}
	}

	groupPVCByName := make(map[string]struct {
		ID      string
		GroupID string
	})
	for _, gs := range groupStorages {
		groupPVCByName[gs.PVCName] = struct {
			ID      string
			GroupID string
		}{ID: gs.ID, GroupID: gs.GroupID}
	}

	var defaultGroupPVCName string
	if usesGroupVolumeTemplate {
		for _, gs := range groupStorages {
			defaultGroupPVCName = gs.PVCName
			break
		}
		if defaultGroupPVCName == "" {
			return "", "", fmt.Errorf("no group storage available for group %s", project.GID)
		}
		groupPVCNames[defaultGroupPVCName] = struct{}{}
	}

	// Bind user storage if needed
	var targetUserPvcName string
	if usesUserVolume {
		safeUsername := k8s.ToSafeK8sName(claims.Username)
		userStorageNs := fmt.Sprintf(config.UserStorageNs, safeUsername)
		targetUserPvcName = fmt.Sprintf(config.UserStoragePVC, safeUsername)
		if err := k8s.MountExistingVolumeToNamespace(userStorageNs, targetUserPvcName, targetNs, targetUserPvcName); err != nil {
			return "", "", fmt.Errorf("failed to bind user storage: %w", err)
		}
	}

	// Bind group storages used in this config
	for pvcName := range groupPVCNames {
		ref, ok := groupPVCByName[pvcName]
		if !ok {
			return "", "", fmt.Errorf("group storage %s not found in group %s", pvcName, project.GID)
		}
		if !claims.IsAdmin && s.Repos != nil && s.Repos.StoragePermission != nil {
			perm, err := s.Repos.StoragePermission.GetPermission(ctx, ref.GroupID, claims.UserID, ref.ID)
			if err != nil || perm == nil || !perm.CanRead() {
				return "", "", fmt.Errorf("user does not have access to group storage %s", pvcName)
			}
		}
		sourceNs := fmt.Sprintf("group-%s-storage", ref.GroupID)
		if err := k8s.MountExistingVolumeToNamespace(sourceNs, pvcName, targetNs, pvcName); err != nil {
			return "", "", fmt.Errorf("failed to bind group storage %s: %w", pvcName, err)
		}
		if defaultGroupPVCName == "" {
			defaultGroupPVCName = pvcName
		}
	}

	return targetUserPvcName, defaultGroupPVCName, nil
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

func collectPVCClaimNames(resources []resource.Resource) map[string]struct{} {
	claims := make(map[string]struct{})
	for _, res := range resources {
		if len(res.ParsedYAML) == 0 {
			continue
		}
		var obj map[string]interface{}
		if err := json.Unmarshal(res.ParsedYAML, &obj); err != nil {
			continue
		}
		podSpecs := findPodSpecs(obj)
		for _, spec := range podSpecs {
			volumes, ok := spec["volumes"].([]interface{})
			if !ok {
				continue
			}
			for _, vol := range volumes {
				vmap, ok := vol.(map[string]interface{})
				if !ok {
					continue
				}
				pvcSource, ok := vmap["persistentVolumeClaim"].(map[string]interface{})
				if !ok {
					continue
				}
				claimName, ok := pvcSource["claimName"].(string)
				if ok && strings.TrimSpace(claimName) != "" {
					claims[claimName] = struct{}{}
				}
			}
		}
	}
	return claims
}
