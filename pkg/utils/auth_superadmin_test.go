package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSuperAdmin(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		isAdmin     bool
		dbErr       bool
		wantErr     bool
		wantAdmin   bool
		description string
	}{
		{
			name:        "super admin user 1",
			userID:      "1",
			isAdmin:     true,
			dbErr:       false,
			wantErr:     false,
			wantAdmin:   true,
			description: "user ID 1 should always be super admin",
		},
		{
			name:        "non-admin user",
			userID:      "2",
			isAdmin:     false,
			dbErr:       false,
			wantErr:     false,
			wantAdmin:   false,
			description: "non-admin user should return false",
		},
		{
			name:        "admin user",
			userID:      "3",
			isAdmin:     true,
			dbErr:       false,
			wantErr:     false,
			wantAdmin:   true,
			description: "admin user should return true",
		},
		{
			name:        "database error",
			userID:      "2",
			isAdmin:     false,
			dbErr:       true,
			wantErr:     true,
			wantAdmin:   false,
			description: "should propagate database errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserGroupRepo{
				superAdmins: map[string]bool{tt.userID: tt.isAdmin},
				shouldErr:   tt.dbErr,
			}

			got, err := IsSuperAdmin(tt.userID, repo)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantAdmin, got, tt.description)
			}
		})
	}
}
