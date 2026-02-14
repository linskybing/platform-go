package form

import (
	"testing"
	"time"
)

// TestFormMessageStructure verifies FormMessage struct
func TestFormMessageStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupMsg func() FormMessage
		verify   func(FormMessage) bool
		scenario string
	}{
		{
			name: "message_minimal",
			setupMsg: func() FormMessage {
				return FormMessage{
					FormID:  "1",
					UserID:  "100",
					Content: "This is a test message",
				}
			},
			verify: func(m FormMessage) bool {
				return m.FormID == "1" && m.UserID == "100" && m.Content != ""
			},
			scenario: "Create minimal form message",
		},
		{
			name: "message_with_id",
			setupMsg: func() FormMessage {
				return FormMessage{
					ID:      "99",
					FormID:  "2",
					UserID:  "101",
					Content: "Message with ID",
				}
			},
			verify: func(m FormMessage) bool {
				return m.ID == "99" && m.FormID == "2"
			},
			scenario: "Message with predefined ID",
		},
		{
			name: "message_with_timestamp",
			setupMsg: func() FormMessage {
				now := time.Now()
				return FormMessage{
					ID:        "1",
					FormID:    "3",
					UserID:    "102",
					Content:   "Timestamped message",
					CreatedAt: now,
				}
			},
			verify: func(m FormMessage) bool {
				return !m.CreatedAt.IsZero()
			},
			scenario: "Message with timestamp",
		},
		{
			name: "message_long_content",
			setupMsg: func() FormMessage {
				longContent := "This is a very long message that spans multiple lines and contains detailed information. It tests the ability of the system to handle longer text content in form messages without truncation or errors."
				return FormMessage{
					FormID:  "4",
					UserID:  "103",
					Content: longContent,
				}
			},
			verify: func(m FormMessage) bool {
				return len(m.Content) > 100
			},
			scenario: "Message with long content",
		},
		{
			name: "message_empty_content_boundary",
			setupMsg: func() FormMessage {
				return FormMessage{
					FormID:  "5",
					UserID:  "104",
					Content: "",
				}
			},
			verify: func(m FormMessage) bool {
				return m.Content == ""
			},
			scenario: "Message with empty content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.setupMsg()
			if !tt.verify(msg) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
