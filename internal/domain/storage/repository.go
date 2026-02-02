package storage

// Repository defines data access interface for storage resources
type Repository interface {
	// PVC operations
	CreatePVC(pvc *PersistentVolumeClaim) error
	GetPVC(id uint) (*PersistentVolumeClaim, error)
	GetPVCByName(namespace, name string) (*PersistentVolumeClaim, error)
	ListPVCs(namespace string) ([]PersistentVolumeClaim, error)
	ListGroupPVCs(groupID uint) ([]PersistentVolumeClaim, error)
	UpdatePVC(pvc *PersistentVolumeClaim) error
	DeletePVC(id uint) error
	DeletePVCByName(namespace, name string) error

	// StorageHub operations
	CreateHub(hub *StorageHub) error
	GetHub(id uint) (*StorageHub, error)
	GetHubByName(namespace, name string) (*StorageHub, error)
	ListHubs(namespace string) ([]StorageHub, error)
	UpdateHub(hub *StorageHub) error
	DeleteHub(id uint) error
	DeleteHubByName(namespace, name string) error
}
