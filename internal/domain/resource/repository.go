package resource

// Repository defines data access interface for resources
type Repository interface {
	CreateResource(resource *Resource) error
	GetResourceByID(rid string) (*Resource, error)
	UpdateResource(resource *Resource) error
	DeleteResource(rid string) error
	ListResourcesByProjectID(pid string) ([]Resource, error)
	ListResourcesByConfigFileID(cfID string) ([]Resource, error)
	GetResourceByConfigFileIDAndName(cfID string, name string) (*Resource, error)
	GetGroupIDByResourceID(rID string) (string, error)
}
