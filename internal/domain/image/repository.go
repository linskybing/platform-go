package image

import "gorm.io/gorm"

type Repository interface {
	FindOrCreateRepository(repo *ContainerRepository) error
	FindOrCreateTag(tag *ContainerTag) error
	GetTagByDigest(repoID string, digest string) (*ContainerTag, error)

	CreateRequest(req *ImageRequest) error
	FindRequestByID(id string) (*ImageRequest, error)
	ListRequests(projectID *string, status string) ([]ImageRequest, error)
	UpdateRequest(req *ImageRequest) error

	CreateAllowListRule(rule *ImageAllowList) error
	ListAllowedImages(projectID *string) ([]ImageAllowList, error)
	FindAllowListRule(projectID *string, repoFullName, tagName string) (*ImageAllowList, error)
	CheckImageAllowed(projectID *string, repoFullName string, tagName string) (bool, error)
	DisableAllowListRule(id string) error

	UpdateClusterStatus(status *ClusterImageStatus) error
	GetClusterStatus(tagID string) (*ClusterImageStatus, error)

	WithTx(tx *gorm.DB) Repository
}
