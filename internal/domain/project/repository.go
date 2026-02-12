package project

// Repository defines data access interface for projects
type Repository interface {
	GetProjectByID(id string) (Project, error)
	GetGroupIDByProjectID(pID string) (string, error)
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	DeleteProject(id string) error
	ListProjects() ([]Project, error)
	ListProjectsByGroup(id string) ([]Project, error)
}
