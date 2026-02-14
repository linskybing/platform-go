package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplacePlaceholders(t *testing.T) {
	template := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: {{name}}\nspec:\n  image: {{image}}"
	values := map[string]string{
		"name":  "my-pod",
		"image": "nginx:latest",
	}

	expected := "apiVersion: v1\nkind: Pod\nmetadata:\n  name: my-pod\nspec:\n  image: nginx:latest"
	result := ReplacePlaceholders(template, values)

	assert.Equal(t, expected, result)
}

func TestReplacePlaceholdersInJSON(t *testing.T) {
	jsonStr := `{
		"metadata": {
			"name": "{{name}}"
		},
		"spec": {
			"containers": [{
				"image": "{{image}}"
			}]
		}
	}`

	values := map[string]string{
		"name":  "my-app",
		"image": "nginx:1.25",
	}

	result, err := ReplacePlaceholdersInJSON(jsonStr, values)
	assert.NoError(t, err)
	assert.Contains(t, result, `"name": "my-app"`)
	assert.Contains(t, result, `"image": "nginx:1.25"`)
}
