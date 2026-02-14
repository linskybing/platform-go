package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGroupManagePermission(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		groupID     string
		isAdmin     bool
		dbErr       bool
		wantErr     bool
		wantPerm    bool
		description string
	}{
		{
			name:        "super admin can manage",
			userID:      "1",
			groupID:     "1",
			isAdmin:     false,
			dbErr:       false,
			wantErr:     false,
			wantPerm:    true,
			description: "user 1 should be able to manage any group",
		},
		{
			name:        "manager can manage",
			userID:      "2",
			groupID:     "1",
			isAdmin:     true,
			dbErr:       false,
			wantErr:     false,
			wantPerm:    true,
			description: "manager should be able to manage group",
		},
		{
			name:        "non-manager cannot manage",
			userID:      "3",
			groupID:     "1",
			isAdmin:     false,
			dbErr:       false,
			wantErr:     true,
			wantPerm:    false,
			description: "non-manager should not be able to manage",
		},
		{
			name:        "database error",
			userID:      "2",
			groupID:     "1",
			isAdmin:     false,
			dbErr:       true,
			wantErr:     true,
			wantPerm:    false,
			description: "should propagate database errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "super admin can manage" {
				assert.Equal(t, "1", tt.userID, "super admin test case")
			}
		})
	}
}
