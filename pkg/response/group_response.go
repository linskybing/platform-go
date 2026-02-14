package response

import "github.com/linskybing/platform-go/internal/domain/group"

// GroupResponse wraps a group payload with a message.
type GroupResponse struct {
	Message string      `json:"message"`
	Group   group.Group `json:"group"`
}
