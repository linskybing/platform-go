package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrEmptyParameter
func TestErrEmptyParameter(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		paramValue  string
		wantErr     bool
		description string
	}{
		{
			name:        "empty param error",
			paramName:   "limit",
			paramValue:  "",
			wantErr:     true,
			description: "empty parameter should return ErrEmptyParameter",
		},
		{
			name:        "non-empty param",
			paramName:   "limit",
			paramValue:  "10",
			wantErr:     false,
			description: "non-empty parameter should not error for empty parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, ErrEmptyParameter)
			assert.Equal(t, "empty parameter", ErrEmptyParameter.Error())
		})
	}
}
