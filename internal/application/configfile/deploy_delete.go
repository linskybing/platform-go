package configfile

import (
	"context"
	"log/slog"

	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/types"
)

func (s *ConfigFileService) DeleteInstance(ctx context.Context, id string, claims *types.Claims) error {
	data, err := s.Repos.Resource.ListResourcesByCommitID(ctx, id)
	if err != nil {
		return err
	}
	commit, err := s.Repos.ConfigFile.GetCommit(ctx, id)
	if err != nil {
		return err
	}

	safeUsername := k8s.ToSafeK8sName(claims.Username)
	ns := k8s.FormatNamespaceName(commit.ProjectID, safeUsername)

	for _, val := range data {
		if err := k8s.DeleteByJson(val.ParsedYAML, ns); err != nil {
			// Continue deleting other resources even if one fails
			slog.Error("failed to delete resource",
				"resource", val.Name,
				"error", err)
		}
	}
	return nil
}

func (s *ConfigFileService) DeleteConfigFileInstance(id string) error {
	commit, err := s.Repos.ConfigFile.GetCommit(context.Background(), id)
	if err != nil {
		return err
	}

	resources, err := s.Repos.Resource.ListResourcesByCommitID(context.Background(), id)
	if err != nil {
		return err
	}

	users, err := s.Repos.User.ListUsersByProjectID(context.Background(), commit.ProjectID)
	if err != nil {
		return err
	}

	for _, user := range users {
		safeUsername := k8s.ToSafeK8sName(user.Username)
		ns := k8s.FormatNamespaceName(commit.ProjectID, safeUsername)
		for _, res := range resources {
			if err := k8s.DeleteByJson(res.ParsedYAML, ns); err != nil {
				slog.Warn("failed to delete instance for user",
					"username", user.Username,
					"namespace", ns,
					"error", err)
			}
		}
	}

	return nil
}
