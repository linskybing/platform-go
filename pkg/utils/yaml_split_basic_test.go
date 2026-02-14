package utils

import (
	"reflect"
	"testing"
)

func TestSplitYAMLDocuments_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "single document",
			input: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
`,
			expected: []string{
				`apiVersion: v1
kind: ConfigMap
metadata:
  name: config1`,
			},
		},
		{
			name: "multiple documents with --- separator",
			input: `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
`,
			expected: []string{
				`apiVersion: v1
kind: ConfigMap
metadata:
  name: config1`,
				`apiVersion: v1
kind: ConfigMap
metadata:
  name: config2`,
			},
		},
		{
			name: "trailing separator",
			input: `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
`,
			expected: []string{
				`apiVersion: v1
kind: ConfigMap
metadata:
  name: config1`,
			},
		},
		{
			name:     "empty input",
			input:    ``,
			expected: []string{},
		},
		{
			name:     "only separators",
			input:    `---`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SplitYAMLDocuments(tt.input)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("SplitYAMLDocuments() = %v, want %v", actual, tt.expected)
			}
		})
	}
}
