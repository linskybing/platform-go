package resource

// Repository defines data access interface for resources
type Repository interface {
	CreateResource(resource *Resource) error
	GetResourceByID(rid string) (*Resource, error)
	UpdateResource(resource *Resource) error
	DeleteResource(rid string) error
	ListResourcesByProjectID(pid string) ([]Resource, error)
	ListResourcesByCommitID(commitID string) ([]Resource, error)
	GetResourceByCommitIDAndName(commitID string, name string) (*Resource, error)
	GetGroupIDByResourceID(rID string) (string, error)
}
