package form

// AllowedTags defines the valid tag values for form submissions.
var AllowedTags = []string{
	"feature",
	"bug",
	"question",
	"resource",
	"access",
	"other",
}

// IsValidTag checks whether the given tag is in the allowed set.
// An empty tag is considered valid (optional field).
func IsValidTag(tag string) bool {
	if tag == "" {
		return true
	}
	for _, t := range AllowedTags {
		if t == tag {
			return true
		}
	}
	return false
}

type CreateFormDTO struct {
	ProjectID   *string `json:"project_id"`
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Tag         string  `json:"tag" binding:"omitempty,oneof=feature bug question resource access other"`
}

type UpdateFormStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=Pending Processing Completed Rejected"`
}

type CreateFormMessageDTO struct {
	Content string `json:"content" binding:"required"`
}
