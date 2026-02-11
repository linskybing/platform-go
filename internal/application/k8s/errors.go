package k8s

import "errors"

var (
	ErrNilRepository = errors.New("repository is required")
	ErrNilRequest    = errors.New("request is required")
	ErrMissingField  = errors.New("required field is missing")

	ErrInvalidID    = errors.New("invalid ID: must be greater than 0")
	ErrInvalidInput = errors.New("invalid input provided")

	ErrDeprecated = errors.New("this feature is deprecated")

	// Storage permission errors
	ErrPermissionDenied = errors.New("permission denied: you don't have access to this storage")
	ErrOnlyAdminsCanSet = errors.New("only group admins can set permissions")
	ErrOnlyAdminsPolicy = errors.New("only group admins can set access policies")
)
