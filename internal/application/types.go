package application

import (
	"github.com/linskybing/platform-go/internal/application/audit"
	"github.com/linskybing/platform-go/internal/application/configfile"
	"github.com/linskybing/platform-go/internal/application/group"
	"github.com/linskybing/platform-go/internal/application/image"
	"github.com/linskybing/platform-go/internal/application/project"
	"github.com/linskybing/platform-go/internal/application/user"
)

// Type aliases for backward compatibility with existing handlers and services
type (
	AuditService      = audit.AuditService
	ConfigFileService = configfile.ConfigFileService
	GroupService      = group.GroupService
	ImageService      = image.ImageService
	ProjectService    = project.ProjectService
	UserService       = user.UserService
	UserGroupService  = group.UserGroupService
)

// Error exports from subpackages
var (
	ErrProjectNotFound     = project.ErrProjectNotFound
	ErrUserNotFound        = user.ErrUserNotFound
	ErrUsernameTaken       = user.ErrUsernameTaken
	ErrMissingOldPassword  = user.ErrMissingOldPassword
	ErrIncorrectPassword   = user.ErrIncorrectPassword
	ErrReservedUser        = user.ErrReservedAdminUser
	ErrConfigFileNotFound  = configfile.ErrConfigFileNotFound
	ErrNoValidYAMLDocument = configfile.ErrNoValidYAMLDocument
	ErrReservedGroupName   = group.ErrReservedGroupName
)

// Other types from image package
type PullJobStatus = image.PullJobStatus
