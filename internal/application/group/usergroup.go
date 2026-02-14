package group

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/utils"
)

var (
	ErrReservedUser                = errors.New("cannot modify reserved user & group 'admin & super'")
	ErrCannotRemoveAdminFromSuper  = errors.New("cannot remove admin user from " + config.ReservedGroupName + " group")
	ErrCannotDowngradeAdminInSuper = errors.New("cannot downgrade admin user role in " + config.ReservedGroupName + " group")
)

type UserGroupService struct {
	Repos *repository.Repos
}

func NewUserGroupService(repos *repository.Repos) *UserGroupService {
	return &UserGroupService{
		Repos: repos,
	}
}

func (s *UserGroupService) AllocateGroupResource(gid string, userName string) error {
	ctx := context.Background()
	projects, err := s.Repos.Project.ListProjectsByGroup(ctx, gid)
	if err != nil {
		return fmt.Errorf("failed to list projects for group %s: %w", gid, err)
	}

	safeUsername := k8s.ToSafeK8sName(userName)

	for _, project := range projects {
		ns := k8s.FormatNamespaceName(project.ID, safeUsername)

		slog.Info("ensuring namespace exists for user",
			"group_id", gid,
			"user", userName,
			"namespace", ns)
		if err := k8s.EnsureNamespaceExists(ns); err != nil {
			slog.Error("failed to create namespace for user",
				"group_id", gid,
				"user", userName,
				"namespace", ns,
				"error", err)
			continue
		}

	}

	return nil
}

func (s *UserGroupService) RemoveGroupResource(gid string, userName string) error {
	ctx := context.Background()
	projects, err := s.Repos.Project.ListProjectsByGroup(ctx, gid)
	if err != nil {
		return fmt.Errorf("failed to list projects for group %s: %w", gid, err)
	}

	safeUsername := k8s.ToSafeK8sName(userName)
	var lastErr error

	for _, project := range projects {
		ns := k8s.FormatNamespaceName(project.ID, safeUsername)

		slog.Info("removing resource namespace for user",
			"group_id", gid,
			"user", userName,
			"namespace", ns)

		if err := k8s.DeleteNamespace(ns); err != nil {
			slog.Warn("failed to delete namespace for user",
				"group_id", gid,
				"user", userName,
				"namespace", ns,
				"error", err)
			lastErr = err
			continue
		}
	}

	return lastErr
}

func (s *UserGroupService) CreateUserGroup(c *gin.Context, userGroup *group.UserGroup) (*group.UserGroup, error) {
	if userGroup == nil {
		return nil, errors.New("user group payload is nil")
	}
	ctx := c.Request.Context()
	projects, err := s.Repos.Project.ListProjectsByGroup(ctx, userGroup.GroupID)
	if err != nil {
		return nil, err
	}
	for _, proj := range projects {
		if proj.MaxProjectUsers <= 0 {
			continue
		}
		count, err := s.Repos.UserGroup.CountUsersByGID(ctx, userGroup.GroupID)
		if err != nil {
			return nil, err
		}
		if count >= int64(proj.MaxProjectUsers) {
			return nil, fmt.Errorf("project %s user limit reached", proj.ID)
		}
	}

	if err := s.Repos.UserGroup.CreateUserGroup(ctx, userGroup); err != nil {
		return nil, err
	}

	uesrName, err := s.Repos.User.GetUsernameByID(ctx, userGroup.UserID)

	if err != nil {
		return nil, err
	}

	if err := s.AllocateGroupResource(userGroup.GroupID, uesrName); err != nil {
		return nil, err
	}

	utils.LogAuditWithConsole(c, "create", "user_group",
		fmt.Sprintf("user_id=%s,group_id=%s", userGroup.UserID, userGroup.GroupID),
		nil, *userGroup, "", s.Repos.Audit)

	return userGroup, nil
}

func (s *UserGroupService) UpdateUserGroup(c *gin.Context, userGroup *group.UserGroup, existing group.UserGroup) (*group.UserGroup, error) {
	ctx := c.Request.Context()
	// Check if trying to downgrade admin user in super group
	groupData, err := s.Repos.Group.GetGroupByID(ctx, userGroup.GroupID)
	if err == nil && groupData.Name == config.ReservedGroupName {
		username, err := s.Repos.User.GetUsernameByID(ctx, userGroup.UserID)
		if err == nil && username == config.ReservedAdminUsername {
			// Check if role is being changed to something other than admin
			if userGroup.Role != "admin" && existing.Role == "admin" {
				return nil, ErrCannotDowngradeAdminInSuper
			}
		}
	}

	if err := s.Repos.UserGroup.UpdateUserGroup(ctx, userGroup); err != nil {
		return nil, err
	}

	utils.LogAuditWithConsole(c, "update", "user_group",
		fmt.Sprintf("user_id=%s,group_id=%s", userGroup.UserID, userGroup.GroupID),
		existing, *userGroup, "", s.Repos.Audit)

	return userGroup, nil
}

func (s *UserGroupService) DeleteUserGroup(c *gin.Context, uid, gid string) error {
	ctx := c.Request.Context()
	oldUserGroup, err := s.Repos.UserGroup.GetUserGroup(ctx, uid, gid)
	if err != nil {
		return err
	}

	// Check if trying to remove admin user from super group
	groupData, err := s.Repos.Group.GetGroupByID(ctx, gid)
	if err == nil && groupData.Name == config.ReservedGroupName {
		username, err := s.Repos.User.GetUsernameByID(ctx, uid)
		if err == nil && username == config.ReservedAdminUsername {
			return ErrCannotRemoveAdminFromSuper
		}
	}

	slog.Debug("removing user from group", "user_id", uid, "group_id", gid)
	if err := s.Repos.UserGroup.DeleteUserGroup(ctx, uid, gid); err != nil {
		return err
	}

	uesrName, err := s.Repos.User.GetUsernameByID(ctx, uid)
	if err != nil {
		return err
	}

	if err := s.RemoveGroupResource(gid, uesrName); err != nil {
		return err
	}

	utils.LogAuditWithConsole(c, "delete", "user_group",
		fmt.Sprintf("user_id=%s,group_id=%s", uid, gid),
		*oldUserGroup, nil, "", s.Repos.Audit)

	return nil
}

func (s *UserGroupService) GetUserGroup(uid, gid string) (group.UserGroup, error) {
	ctx := context.Background()
	ug, err := s.Repos.UserGroup.GetUserGroup(ctx, uid, gid)
	if err != nil {
		return group.UserGroup{}, err
	}
	return *ug, nil
}

func (s *UserGroupService) GetUserGroupsByUID(uid string) ([]group.UserGroup, error) {
	return s.Repos.UserGroup.GetUserGroupsByUID(context.Background(), uid)
}

func (s *UserGroupService) GetUserGroupsByGID(gid string) ([]group.UserGroup, error) {
	return s.Repos.UserGroup.GetUserGroupsByGID(context.Background(), gid)
}

func (s *UserGroupService) FormatByUID(records []group.UserGroup) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	ctx := context.Background()

	for _, r := range records {
		// Get group name for this group
		groupData, err := s.Repos.Group.GetGroupByID(ctx, r.GroupID)
		groupName := ""
		if err == nil {
			groupName = groupData.Name
		}

		groupInfo := map[string]interface{}{
			"group_id":   r.GroupID,
			"group_name": groupName,
			"role":       r.Role,
		}

		if u, exists := result[r.UserID]; exists {
			// Append to existing groups array
			groups := u["groups"].([]map[string]interface{})
			groups = append(groups, groupInfo)
			u["groups"] = groups
		} else {
			// Get username
			username, err := s.Repos.User.GetUsernameByID(ctx, r.UserID)
			if err != nil {
				username = "" // If we can't get the username, use empty string
			}

			// Create new entry with groups array
			result[r.UserID] = map[string]interface{}{
				"user_id":   r.UserID,
				"user_name": username,
				"groups":    []map[string]interface{}{groupInfo},
			}
		}
	}
	return result
}

func (s *UserGroupService) FormatByGID(records []group.UserGroup) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	ctx := context.Background()

	for _, r := range records {
		// Use preloaded username from User relationship
		username := r.User.Username
		if username == "" {
			username = "" // If not loaded, use empty string
		}

		userInfo := map[string]interface{}{
			"user_id":  r.UserID,
			"username": username,
			"role":     r.Role,
		}

		if g, exists := result[r.GroupID]; exists {
			// Append to existing users array
			users := g["users"].([]map[string]interface{})
			users = append(users, userInfo)
			g["users"] = users
		} else {
			// Get group name
			groupData, err := s.Repos.Group.GetGroupByID(ctx, r.GroupID)
			groupName := ""
			if err == nil {
				groupName = groupData.Name
			}

			// Create new entry with users array
			result[r.GroupID] = map[string]interface{}{
				"group_id":   r.GroupID,
				"group_name": groupName,
				"users":      []map[string]interface{}{userInfo},
			}
		}
	}
	return result
}
