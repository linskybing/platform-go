package response

import (
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
)

// TestGroupResponseStructure tests GroupResponse struct
func TestGroupResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response GroupResponse
		verify   func(GroupResponse) bool
		scenario string
	}{
		{
			name: "group_response_with_data",
			response: GroupResponse{
				Message: "Group retrieved successfully",
				Group: group.Group{
					GroupName: "TestGroup",
				},
			},
			verify: func(r GroupResponse) bool {
				return r.Message != "" && r.Group.GroupName == "TestGroup"
			},
			scenario: "Group response with group data",
		},
		{
			name: "group_response_empty_name",
			response: GroupResponse{
				Message: "Group created",
				Group: group.Group{
					GroupName: "",
				},
			},
			verify: func(r GroupResponse) bool {
				return r.Message != ""
			},
			scenario: "Group response with empty name",
		},
		{
			name: "group_response_long_message",
			response: GroupResponse{
				Message: "Group operation completed successfully with all validations passed",
				Group: group.Group{
					GroupName: "LongMessageGroup",
				},
			},
			verify: func(r GroupResponse) bool {
				return len(r.Message) > 30
			},
			scenario: "Group response with long message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.response) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
