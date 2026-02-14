package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/linskybing/platform-go/internal/domain/common"
	"github.com/linskybing/platform-go/internal/domain/project"
	"gorm.io/gorm"
)

type ProjectRepo interface {
	CreateProject(ctx context.Context, node *project.Project) error
	CreateNode(ctx context.Context, node *project.Project) error
	UpdateProject(ctx context.Context, node *project.Project) error
	UpdateNode(ctx context.Context, node *project.Project) error
	GetProjectByID(ctx context.Context, id string) (*project.Project, error)
	GetNode(ctx context.Context, id string) (*project.Project, error)
	GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error)
	ListProjects(ctx context.Context) ([]project.Project, error)
	ListNodes(ctx context.Context, parentID *string) ([]project.Project, error)
	ListProjectsByGroup(ctx context.Context, gid string) ([]project.Project, error)
	ListDescendantProjects(ctx context.Context, groupOwnerIDs []string) ([]project.Project, error)

	GetSubtree(ctx context.Context, rootID string) ([]project.Project, error)
	GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error)
	MoveNode(ctx context.Context, nodeID, newParentID string) error
	DeleteNode(ctx context.Context, id string) error

	CreateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error
	UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error
	GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error)

	WithTx(tx *gorm.DB) ProjectRepo
}

type ProjectRepoImpl struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) ProjectRepo {
	return &ProjectRepoImpl{db: db}
}

func (r *ProjectRepoImpl) CreateProject(ctx context.Context, node *project.Project) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *ProjectRepoImpl) CreateNode(ctx context.Context, node *project.Project) error {
	return r.CreateProject(ctx, node)
}

func (r *ProjectRepoImpl) UpdateProject(ctx context.Context, node *project.Project) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *ProjectRepoImpl) UpdateNode(ctx context.Context, node *project.Project) error {
	return r.UpdateProject(ctx, node)
}

func (r *ProjectRepoImpl) GetProjectByID(ctx context.Context, id string) (*project.Project, error) {
	var n project.Project
	err := r.db.WithContext(ctx).Preload("ResourcePlan").First(&n, "p_id = ?", id).Error
	return &n, err
}

func (r *ProjectRepoImpl) GetNode(ctx context.Context, id string) (*project.Project, error) {
	return r.GetProjectByID(ctx, id)
}

func (r *ProjectRepoImpl) GetNodeByOwner(ctx context.Context, ownerID string) (*project.Project, error) {
	var n project.Project
	err := r.db.WithContext(ctx).First(&n, "owner_id = ?", ownerID).Error
	return &n, err
}

func (r *ProjectRepoImpl) ListProjects(ctx context.Context) ([]project.Project, error) {
	var nodes []project.Project
	err := r.db.WithContext(ctx).Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) ListNodes(ctx context.Context, parentID *string) ([]project.Project, error) {
	var nodes []project.Project
	query := r.db.WithContext(ctx).Preload("ResourcePlan")
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	err := query.Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) ListProjectsByGroup(ctx context.Context, gid string) ([]project.Project, error) {
	var nodes []project.Project
	err := r.db.WithContext(ctx).Where("owner_id = ?", gid).Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) GetSubtree(ctx context.Context, rootID string) ([]project.Project, error) {
	var nodes []project.Project
	subQuery := r.db.Model(&project.Project{}).Select("path").Where("p_id = ?", rootID)
	err := r.db.WithContext(ctx).Preload("ResourcePlan").Where("path <@ (?)", subQuery).Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) GetAncestors(ctx context.Context, nodeID string) ([]project.Project, error) {
	var nodes []project.Project
	subQuery := r.db.Model(&project.Project{}).Select("path").Where("p_id = ?", nodeID)
	err := r.db.WithContext(ctx).Preload("ResourcePlan").Where("path @> (?)", subQuery).Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) MoveNode(ctx context.Context, nodeID, newParentID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var node project.Project
		if err := tx.First(&node, "p_id = ?", nodeID).Error; err != nil {
			return err
		}

		var newParent project.Project
		if err := tx.First(&newParent, "p_id = ?", newParentID).Error; err != nil {
			return err
		}

		oldPath := node.Path
		newNodeID := strings.ReplaceAll(node.ID, "-", "_")
		newPath := common.Ltree(fmt.Sprintf("%s.%s", newParent.Path, newNodeID))

		// Update the node itself
		if err := tx.Model(&node).Updates(map[string]interface{}{
			"parent_id": newParentID,
			"path":      newPath,
		}).Error; err != nil {
			return err
		}

		// Update all descendants
		// path = new_path || subpath(path, nlevel(old_path))
		query := `UPDATE projects SET path = ? || subpath(path, nlevel(?)) WHERE path <@ ? AND p_id != ?`
		return tx.Exec(query, newPath, oldPath, oldPath, nodeID).Error
	})
}

func (r *ProjectRepoImpl) DeleteNode(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&project.Project{}, "p_id = ?", id).Error
}

func (r *ProjectRepoImpl) ListDescendantProjects(ctx context.Context, groupOwnerIDs []string) ([]project.Project, error) {
	var nodes []project.Project
	if len(groupOwnerIDs) == 0 {
		return nodes, nil
	}
	query := `SELECT * FROM projects WHERE path <@ ANY (SELECT path FROM projects WHERE owner_id IN ?)`
	err := r.db.WithContext(ctx).Raw(query, groupOwnerIDs).Scan(&nodes).Error
	return nodes, err
}

func (r *ProjectRepoImpl) CreateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *ProjectRepoImpl) UpdateResourcePlan(ctx context.Context, plan *project.ResourcePlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *ProjectRepoImpl) GetResourcePlan(ctx context.Context, projectID string) (*project.ResourcePlan, error) {
	var p project.ResourcePlan
	err := r.db.WithContext(ctx).First(&p, "project_id = ?", projectID).Error
	return &p, err
}

func (r *ProjectRepoImpl) WithTx(tx *gorm.DB) ProjectRepo {
	if tx == nil {
		return r
	}
	return &ProjectRepoImpl{db: tx}
}
