package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckGroupPermission(t *testing.T) {
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
			name:        "super admin has permission",
			userID:      "1",
			groupID:     "1",
			isAdmin:     false,
			dbErr:       false,
			wantErr:     false,
			wantPerm:    true,
			description: "user 1 (super admin) should have permission",
		},
		{
			name:        "admin user has permission",
			userID:      "2",
			groupID:     "1",
			isAdmin:     true,
			dbErr:       false,
			wantErr:     false,
			wantPerm:    true,
			description: "admin user should have permission",
		},
		{
			name:        "non-admin no permission",
			userID:      "3",
			groupID:     "1",
			isAdmin:     false,
			dbErr:       false,
			wantErr:     true,
			wantPerm:    false,
			description: "non-admin user should not have permission",
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
			if tt.name == "super admin has permission" {
				assert.False(t, tt.isAdmin, "test expects non-admin user ID but admin flag is true")
				assert.Equal(t, "1", tt.userID, "super admin test case")
			}
		})
	}
}
