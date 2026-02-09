package gpu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertGPUToMPSAndBack(t *testing.T) {
	tests := []struct {
		g int
	}{
		{g: 0}, {g: 1}, {g: 2}, {g: 15},
	}

	for _, tt := range tests {
		m := ConvertGPUToMPS(tt.g)
		got := ConvertMPSToGPU(m)
		assert.Equal(t, tt.g, got)
	}
}

func TestValidateGPUQuota(t *testing.T) {
	assert.False(t, ValidateGPUQuota(0))
	assert.False(t, ValidateGPUQuota(-1))
	assert.True(t, ValidateGPUQuota(1))
}

func TestCalculateMPSUsageAndQuota(t *testing.T) {
	reqs := []int{10, 20, 0}
	total := CalculateMPSUsage(reqs)
	assert.Equal(t, 30, total)
	assert.True(t, IsWithinQuota(10, 20))
	assert.False(t, IsWithinQuota(30, 20))
}
