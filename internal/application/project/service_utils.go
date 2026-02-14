package project

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/linskybing/platform-go/pkg/k8s"
)

// RemoveProjectResources cleans up Kubernetes namespaces and group storage.
func (s *ProjectService) RemoveProjectResources(projectID string) error {
	p, err := s.Repos.Project.GetNode(context.Background(), projectID)
	if err != nil {
		return err
	}

	if p.ParentID == nil {
		slog.Warn("project has no parent, skipping user namespace cleanup", "project_id", projectID)
		return nil
	}

	// Resolve Parent Group ID
	parentNode, err := s.Repos.Project.GetNode(context.Background(), *p.ParentID)
	if err != nil {
		return fmt.Errorf("failed to get parent node: %w", err)
	}

	if parentNode.OwnerID == nil {
		slog.Warn("parent node has no owner (group), skipping user cleanup", "node_id", parentNode.ID)
	} else {
		users, err := s.Repos.Group.ListUsersInGroup(context.Background(), *parentNode.OwnerID)
		if err != nil {
			return err
		}

		for _, u := range users {
			ns := k8s.FormatNamespaceName(p.ID, k8s.ToSafeK8sName(u.Username))
			if err := k8s.DeleteNamespace(ns); err != nil {
				slog.Error("failed to delete ns", "ns", ns, "error", err)
			}
		}
	}

	groupStorageNs := k8s.GenerateSafeResourceName("group", p.Name, p.ID)
	return k8s.DeleteNamespace(groupStorageNs)
}

func (s *ProjectService) invalidateProjectCache(projectID string) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, projectListKey(), projectByIDKey(projectID))
	_ = s.cache.InvalidatePrefix(ctx, "cache:project:by-user:")
}

func projectListKey() string                { return "cache:project:list" }
func projectByIDKey(id string) string       { return fmt.Sprintf("cache:project:by-id:%s", id) }
func projectByUserKey(userID string) string { return fmt.Sprintf("cache:project:by-user:%s", userID) }
