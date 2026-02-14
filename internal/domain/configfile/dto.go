package configfile

type ConfigFileUpdateDTO struct {
	RawYaml *string `form:"raw_yaml"`
	Message *string `form:"message"`
}

type CreateConfigFileInput struct {
	RawYaml   string `form:"raw_yaml" binding:"required"`
	ProjectID string `form:"project_id" binding:"required"`
	Message   string `form:"message"`
}

// GetProjectID returns the project ID for GID lookup
func (d CreateConfigFileInput) GetProjectID() string {
	return d.ProjectID
}

type ProjectGetter interface {
	GetGroupIDByProjectID(projectID string) string
}
