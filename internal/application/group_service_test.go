package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/group"
)

func TestGroupServiceCRUD(t *testing.T) {
	svc, stubGroup, c := setupGroupService(t)

	t.Run("ListGroups success", func(t *testing.T) {
		stubGroup.list = func(ctx context.Context) ([]group.Group, error) {
			return []group.Group{
				{ID: "1", GID: "1", Name: "dev", GroupName: "dev"},
				{ID: "2", GID: "2", Name: "ops", GroupName: "ops"},
			}, nil
		}

		groups, err := svc.ListGroups()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(groups) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(groups))
		}
	})

	t.Run("GetGroup success", func(t *testing.T) {
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return &group.Group{ID: "1", GID: "1", Name: "dev", GroupName: "dev"}, nil
		}
		group, err := svc.GetGroup("1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "dev" {
			t.Fatalf("expected dev, got %s", group.GroupName)
		}
	})

	t.Run("GetGroup not found", func(t *testing.T) {
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return nil, errors.New("not found")
		}
		_, err := svc.GetGroup("99")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("CreateGroup success", func(t *testing.T) {
		input := group.GroupCreateDTO{GroupName: "qa", Description: nil}
		stubGroup.create = func(ctx context.Context, g *group.Group) error {
			return nil
		}

		group, err := svc.CreateGroup(c, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "qa" {
			t.Fatalf("expected qa, got %s", group.GroupName)
		}
	})

	t.Run("CreateGroup reserved name", func(t *testing.T) {
		input := group.GroupCreateDTO{GroupName: "super"}
		_, err := svc.CreateGroup(c, input)
		if !errors.Is(err, application.ErrReservedGroupName) {
			t.Fatalf("expected reserved group name error, got %v", err)
		}
	})

	t.Run("UpdateGroup success", func(t *testing.T) {
		old := &group.Group{ID: "1", GID: "1", Name: "dev", GroupName: "dev"}
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return old, nil
		}
		stubGroup.update = func(ctx context.Context, g *group.Group) error {
			return nil
		}

		newName := "devops"
		newDesc := "updated description"
		input := group.GroupUpdateDTO{GroupName: &newName, Description: &newDesc}
		group, err := svc.UpdateGroup(c, "1", input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if group.GroupName != "devops" || group.Description != "updated description" {
			t.Fatalf("unexpected updated group %+v", group)
		}
	})

	t.Run("UpdateGroup reserved name", func(t *testing.T) {
		old := &group.Group{ID: "1", GID: "1", Name: "dev", GroupName: "dev"}
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return old, nil
		}
		newName := "super"
		input := group.GroupUpdateDTO{GroupName: &newName}
		_, err := svc.UpdateGroup(c, "1", input)
		if !errors.Is(err, application.ErrReservedGroupName) {
			t.Fatalf("expected reserved group name error, got %v", err)
		}
	})

	t.Run("UpdateGroup not found", func(t *testing.T) {
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return nil, errors.New("not found")
		}
		newName := "newname"
		input := group.GroupUpdateDTO{GroupName: &newName}
		_, err := svc.UpdateGroup(c, "99", input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("DeleteGroup success", func(t *testing.T) {
		existing := &group.Group{ID: "1", GID: "1", Name: "dev", GroupName: "dev"}
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return existing, nil
		}
		stubGroup.deleteFunc = func(ctx context.Context, id string) error {
			return nil
		}

		err := svc.DeleteGroup(c, "1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("DeleteGroup not found", func(t *testing.T) {
		stubGroup.get = func(ctx context.Context, id string) (*group.Group, error) {
			return nil, errors.New("not found")
		}
		err := svc.DeleteGroup(c, "99")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
