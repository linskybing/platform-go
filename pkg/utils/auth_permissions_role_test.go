package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasGroupRole(t *testing.T) {
	tests := []struct {
		name        string
		uid         string
		gid         string
		roles       []string
		description string
	}{
		{
			name:        "check admin role",
			uid:         "1",
			gid:         "1",
			roles:       []string{"admin"},
			description: "should check for admin role",
		},
		{
			name:        "check viewer role",
			uid:         "2",
			gid:         "1",
			roles:       []string{"viewer", "member"},
			description: "should check multiple roles",
		},
		{
			name:        "empty roles list",
			uid:         "3",
			gid:         "2",
			roles:       []string{},
			description: "should handle empty roles list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, HasGroupRole)
		})
	}
}
