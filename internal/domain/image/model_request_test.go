package image

import (
	"testing"
	"time"
)

func TestImageRequestStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupIR  func() ImageRequest
		verify   func(ImageRequest) bool
		scenario string
	}{
		{
			name: "pending_request",
			setupIR: func() ImageRequest {
				return ImageRequest{
					UserID:         "1",
					ProjectID:      ptrString("50"),
					InputRegistry:  "docker.io",
					InputImageName: "ubuntu",
					InputTag:       "latest",
					Status:         "pending",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "pending" && ir.InputRegistry == "docker.io"
			},
			scenario: "Create pending image request",
		},
		{
			name: "approved_request",
			setupIR: func() ImageRequest {
				reviewedAt := time.Now()
				return ImageRequest{
					UserID:         "2",
					ProjectID:      nil,
					InputRegistry:  "gcr.io",
					InputImageName: "my-app",
					InputTag:       "v1.0",
					Status:         "approved",
					ReviewerID:     ptrString("100"),
					ReviewedAt:     &reviewedAt,
					ReviewerNote:   "Approved for production",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "approved" && ir.ReviewerID != nil
			},
			scenario: "Approved image request with review",
		},
		{
			name: "rejected_request",
			setupIR: func() ImageRequest {
				reviewedAt := time.Now()
				return ImageRequest{
					UserID:         "3",
					ProjectID:      ptrString("51"),
					InputRegistry:  "quay.io",
					InputImageName: "untrusted-image",
					InputTag:       "v0.1",
					Status:         "rejected",
					ReviewerID:     ptrString("101"),
					ReviewedAt:     &reviewedAt,
					ReviewerNote:   "Security concerns detected",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "rejected" && len(ir.ReviewerNote) > 0
			},
			scenario: "Rejected image request",
		},
		{
			name: "unreviewed_request",
			setupIR: func() ImageRequest {
				return ImageRequest{
					UserID:         "4",
					ProjectID:      nil,
					InputRegistry:  "docker.io",
					InputImageName: "postgres",
					InputTag:       "15",
					Status:         "pending",
					ReviewerID:     nil,
					ReviewedAt:     nil,
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.ReviewerID == nil && ir.ReviewedAt == nil
			},
			scenario: "Unreviewed request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupIR()
			if !tt.verify(req) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
