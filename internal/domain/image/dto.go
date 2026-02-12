package image

type CreateImageRequestDTO struct {
	Registry  string  `json:"registry"`
	ImageName string  `json:"image_name" binding:"required"`
	Tag       string  `json:"tag" binding:"required"`
	ProjectID *string `json:"project_id"`
}

type UpdateImageRequestDTO struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Note   string `json:"note"`
}

type ApplyReviewDTO struct {
	Note string `json:"note"`
}

type AllowedImageDTO struct {
	ID        string  `json:"id"`
	Registry  string  `json:"registry"`
	ImageName string  `json:"image_name"`
	Tag       string  `json:"tag"`
	Digest    string  `json:"digest"`
	ProjectID *string `json:"project_id"`
	IsGlobal  bool    `json:"is_global"`
	IsPulled  bool    `json:"is_pulled"`
}

// PullImageRequestDTO represents payload accepted by the pull image endpoint.
type PullImageRequestDTO struct {
	Names []string `json:"names"`
	Name  string   `json:"name"`
	Tag   string   `json:"tag"`
}

type PullRequestDTO struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

// AddProjectImageDTO is used when submitting a project-scoped image request.
type AddProjectImageDTO struct {
	ImageName string `json:"image_name" binding:"required"`
	Tag       string `json:"tag"`
}
