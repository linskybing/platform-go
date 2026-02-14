package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLToJSON(t *testing.T) {
	yamlContent := `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
    - name: nginx
      image: nginx:1.25
`

	expectedContains := `"kind": "Pod"`
	jsonBytes, err := YAMLToJSON([]byte(yamlContent))
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), expectedContains)
}
