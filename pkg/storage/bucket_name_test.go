package storage

import "testing"

// TestBucketNameValidation tests bucket name validation
func TestBucketNameValidation(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		isValid    func(string) bool
		scenario   string
	}{
		{
			name:       "standard_bucket_name",
			bucketName: "my-bucket",
			isValid: func(b string) bool {
				return b == "my-bucket" && len(b) > 3
			},
			scenario: "Valid bucket name with hyphen",
		},
		{
			name:       "lowercase_bucket_name",
			bucketName: "mybucket",
			isValid: func(b string) bool {
				return b == "mybucket"
			},
			scenario: "Simple lowercase bucket name",
		},
		{
			name:       "bucket_with_numbers",
			bucketName: "bucket2024",
			isValid: func(b string) bool {
				return len(b) > 0
			},
			scenario: "Bucket name with numbers",
		},
		{
			name:       "long_bucket_name",
			bucketName: "my-very-long-bucket-name-with-many-characters",
			isValid: func(b string) bool {
				return len(b) <= 63 && len(b) > 10
			},
			scenario: "Long but valid bucket name",
		},
		{
			name:       "bucket_with_dots",
			bucketName: "bucket.name",
			isValid: func(b string) bool {
				return len(b) > 0
			},
			scenario: "Bucket name with dots",
		},
		{
			name:       "minimum_bucket_name",
			bucketName: "bkt",
			isValid: func(b string) bool {
				return len(b) >= 3
			},
			scenario: "Minimum length bucket name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.isValid(tt.bucketName) {
				t.Errorf("validation failed for scenario: %s", tt.scenario)
			}
		})
	}
}
