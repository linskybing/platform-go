package dto

type ConfigFileUpdateDTO struct {
	Filename  *string `form:"filename"`
	MinIOPath *string `form:"minio_path"`
	ProjectID *uint   `form:"project_id"`
}

type CreateConfigFileInput struct {
	RawYaml   string `form:"raw_yaml"`
	ProjectID uint   `form:"project_id"`
}
