package k8s

import "errors"

var (
	ErrNilRepository = errors.New("repository is required")
	ErrNilRequest    = errors.New("request is required")
	ErrMissingField  = errors.New("required field is missing")

	ErrInvalidID    = errors.New("invalid ID: must be greater than 0")
	ErrInvalidInput = errors.New("invalid input provided")

	ErrDeprecated = errors.New("this feature is deprecated")
)
