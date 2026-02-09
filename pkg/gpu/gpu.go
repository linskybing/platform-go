package gpu

// MPSUnitsPerGPU defines how many logical MPS units we treat as one full GPU
const MPSUnitsPerGPU = 10

// ConvertGPUToMPS converts dedicated GPU count to equivalent MPS units
func ConvertGPUToMPS(gpuCount int) int {
	return gpuCount * MPSUnitsPerGPU
}

// ConvertMPSToGPU converts MPS units to equivalent GPU count (round up)
func ConvertMPSToGPU(mpsUnits int) int {
	gpus := mpsUnits / MPSUnitsPerGPU
	if mpsUnits%MPSUnitsPerGPU > 0 {
		gpus++
	}
	return gpus
}

// ValidateGPUQuota returns true if the GPU quota is considered valid (positive)
func ValidateGPUQuota(quota int) bool {
	return quota > 0
}

// CalculateMPSUsage sums a list of MPS requests
func CalculateMPSUsage(requests []int) int {
	total := 0
	for _, r := range requests {
		total += r
	}
	return total
}

// IsWithinQuota checks if requested amount is within available
func IsWithinQuota(requested, available int) bool {
	return requested <= available
}
