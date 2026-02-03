package group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/utils"
)

var ErrReservedGroupName = errors.New("cannot use reserved group name '" + config.ReservedGroupName + "'")

type GroupService struct {
	Repos *repository.Repos
	cache *cache.Service
}

func NewGroupService(repos *repository.Repos) *GroupService {
	return NewGroupServiceWithCache(repos, nil)
}

func NewGroupServiceWithCache(repos *repository.Repos, cacheSvc *cache.Service) *GroupService {
	return &GroupService{
		Repos: repos,
		cache: cacheSvc,
	}
}

const groupCacheTTL = 5 * time.Minute

func (s *GroupService) ListGroups() ([]group.Group, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []group.Group
		if err := s.cache.GetJSON(context.Background(), groupListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	groups, err := s.Repos.Group.GetAllGroups()
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), groupListKey(), groups, groupCacheTTL)
	}

	return groups, nil
}

func (s *GroupService) GetGroup(id uint) (group.Group, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached group.Group
		if err := s.cache.GetJSON(context.Background(), groupByIDKey(id), &cached); err == nil {
			return cached, nil
		}
	}

	grp, err := s.Repos.Group.GetGroupByID(id)
	if err != nil {
		return group.Group{}, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), groupByIDKey(id), grp, groupCacheTTL)
	}
	return grp, nil
}

func (s *GroupService) CreateGroup(c *gin.Context, input group.GroupCreateDTO) (group.Group, error) {
	if input.GroupName == config.ReservedGroupName {
		return group.Group{}, ErrReservedGroupName
	}

	grp := group.Group{
		GroupName: input.GroupName,
	}
	if input.Description != nil {
		grp.Description = *input.Description
	}

	err := s.Repos.Group.CreateGroup(&grp)
	if err != nil {
		return group.Group{}, err
	}
	s.invalidateGroupCache(grp.GID)
	utils.LogAuditWithConsole(c, "create", "group", fmt.Sprintf("g_id=%d", grp.GID), nil, grp, "", s.Repos.Audit)

	return grp, nil
}

func (s *GroupService) UpdateGroup(c *gin.Context, id uint, input group.GroupUpdateDTO) (group.Group, error) {
	grp, err := s.Repos.Group.GetGroupByID(id)
	if err != nil {
		return group.Group{}, err
	}

	// Cannot modify the reserved super group's name
	if grp.GroupName == config.ReservedGroupName && input.GroupName != nil {
		return group.Group{}, ErrReservedGroupName
	}

	oldGroup := grp

	if input.GroupName != nil {
		// Prevent updating reserved group name (both TO and FROM)
		if grp.GroupName == config.ReservedGroupName {
			return group.Group{}, ErrReservedGroupName
		}
		if *input.GroupName == config.ReservedGroupName {
			return group.Group{}, ErrReservedGroupName
		}
		grp.GroupName = *input.GroupName
	}
	if input.Description != nil {
		grp.Description = *input.Description
	}

	err = s.Repos.Group.UpdateGroup(&grp)
	if err != nil {
		return group.Group{}, err
	}
	s.invalidateGroupCache(grp.GID)

	utils.LogAuditWithConsole(c, "update", "group", fmt.Sprintf("g_id=%d", grp.GID), oldGroup, grp, "", s.Repos.Audit)

	return grp, nil
}

func (s *GroupService) DeleteGroup(c *gin.Context, id uint) error {
	group, err := s.Repos.Group.GetGroupByID(id)
	if err != nil {
		return err
	}

	if group.GroupName == config.ReservedGroupName {
		return ErrReservedGroupName
	}

	err = s.Repos.Group.DeleteGroup(id)
	if err != nil {
		return err
	}
	s.invalidateGroupCache(group.GID)

	utils.LogAuditWithConsole(c, "delete", "group", fmt.Sprintf("g_id=%d", group.GID), group, nil, "", s.Repos.Audit)

	return nil
}

func groupListKey() string {
	return "cache:group:list"
}

func groupByIDKey(id uint) string {
	return fmt.Sprintf("cache:group:by-id:%d", id)
}

func (s *GroupService) invalidateGroupCache(id uint) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, groupListKey(), groupByIDKey(id))
}
