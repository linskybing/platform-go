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
	ctx := context.Background()
	if s.cache != nil && s.cache.Enabled() {
		var cached []group.Group
		if err := s.cache.GetJSON(ctx, groupListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	groups, err := s.Repos.Group.List(ctx)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(ctx, groupListKey(), groups, groupCacheTTL)
	}

	return groups, nil
}

func (s *GroupService) GetGroup(id string) (group.Group, error) {
	ctx := context.Background()
	if s.cache != nil && s.cache.Enabled() {
		var cached group.Group
		if err := s.cache.GetJSON(ctx, groupByIDKey(id), &cached); err == nil {
			return cached, nil
		}
	}

	grp, err := s.Repos.Group.Get(ctx, id)
	if err != nil {
		return group.Group{}, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(ctx, groupByIDKey(id), grp, groupCacheTTL)
	}
	return *grp, nil
}

func (s *GroupService) CreateGroup(c *gin.Context, input group.GroupCreateDTO) (group.Group, error) {
	if input.GroupName == config.ReservedGroupName {
		return group.Group{}, ErrReservedGroupName
	}

	// Check for duplicate name
	existing, _ := s.Repos.Group.GetByGroupName(context.Background(), input.GroupName)
	if existing != nil {
		return group.Group{}, errors.New("group already exists")
	}

	grp := group.Group{
		Name: input.GroupName,
	}
	if input.Description != nil {
		grp.Description = *input.Description
	}

	ctx := c.Request.Context()
	err := s.Repos.Group.Create(ctx, &grp)
	if err != nil {
		return group.Group{}, err
	}
	s.invalidateGroupCache(grp.ID)
	utils.LogAuditWithConsole(c, "create", "group", fmt.Sprintf("g_id=%s", grp.ID), nil, grp, "", s.Repos.Audit)

	return grp, nil
}

func (s *GroupService) UpdateGroup(c *gin.Context, id string, input group.GroupUpdateDTO) (group.Group, error) {
	ctx := c.Request.Context()
	grp, err := s.Repos.Group.Get(ctx, id)
	if err != nil {
		return group.Group{}, err
	}

	// Cannot modify the reserved super group's name
	if grp.Name == config.ReservedGroupName && input.GroupName != nil {
		return group.Group{}, ErrReservedGroupName
	}

	oldGroup := *grp

	if input.GroupName != nil {
		if grp.Name == config.ReservedGroupName || *input.GroupName == config.ReservedGroupName {
			return group.Group{}, ErrReservedGroupName
		}
		grp.Name = *input.GroupName
	}
	if input.Description != nil {
		grp.Description = *input.Description
	}

	err = s.Repos.Group.UpdateGroup(ctx, grp)
	if err != nil {
		return group.Group{}, err
	}
	s.invalidateGroupCache(grp.ID)

	utils.LogAuditWithConsole(c, "update", "group", fmt.Sprintf("g_id=%s", grp.ID), oldGroup, *grp, "", s.Repos.Audit)

	return *grp, nil
}

func (s *GroupService) DeleteGroup(c *gin.Context, id string) error {
	ctx := c.Request.Context()
	grp, err := s.Repos.Group.Get(ctx, id)
	if err != nil {
		return err
	}

	if grp.Name == config.ReservedGroupName {
		return ErrReservedGroupName
	}

	err = s.Repos.Group.Delete(ctx, id)
	if err != nil {
		return err
	}
	s.invalidateGroupCache(grp.ID)

	utils.LogAuditWithConsole(c, "delete", "group", fmt.Sprintf("g_id=%s", grp.ID), *grp, nil, "", s.Repos.Audit)

	return nil
}

func groupListKey() string {
	return "cache:group:list"
}

func groupByIDKey(id string) string {
	return fmt.Sprintf("cache:group:by-id:%s", id)
}

func (s *GroupService) invalidateGroupCache(id string) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, groupListKey(), groupByIDKey(id))
}
