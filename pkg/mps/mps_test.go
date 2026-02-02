package mps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConvertGPUToMPS - Table-driven tests
func TestConvertGPUToMPS(t *testing.T) {
	tests := []struct {
		name        string
		gpuCount    int
		wantMPS     int
		description string
	}{
		{
			name:        "zero GPUs",
			gpuCount:    0,
			wantMPS:     0,
			description: "0 GPUs should convert to 0 MPS units",
		},
		{
			name:        "one GPU",
			gpuCount:    1,
			wantMPS:     10,
			description: "1 GPU should convert to 10 MPS units",
		},
		{
			name:        "two GPUs",
			gpuCount:    2,
			wantMPS:     20,
			description: "2 GPUs should convert to 20 MPS units",
		},
		{
			name:        "large GPU count",
			gpuCount:    100,
			wantMPS:     1000,
			description: "100 GPUs should convert to 1000 MPS units",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertGPUToMPS(tt.gpuCount)
			assert.Equal(t, tt.wantMPS, got, tt.description)
		})
	}
}

// TestConvertMPSToGPU - Table-driven tests with rounding up
func TestConvertMPSToGPU(t *testing.T) {
	tests := []struct {
		name        string
		mpsUnits    int
		wantGPU     int
		description string
	}{
		{
			name:        "zero MPS units",
			mpsUnits:    0,
			wantGPU:     0,
			description: "0 MPS units should convert to 0 GPUs",
		},
		{
			name:        "exact GPU units",
			mpsUnits:    10,
			wantGPU:     1,
			description: "10 MPS units should convert to 1 GPU",
		},
		{
			name:        "exact multiple GPUs",
			mpsUnits:    30,
			wantGPU:     3,
			description: "30 MPS units should convert to 3 GPUs",
		},
		{
			name:        "partial GPU rounding up",
			mpsUnits:    15,
			wantGPU:     2,
			description: "15 MPS units should round up to 2 GPUs",
		},
		{
			name:        "one unit rounding up",
			mpsUnits:    1,
			wantGPU:     1,
			description: "1 MPS unit should round up to 1 GPU",
		},
		{
			name:        "large MPS units",
			mpsUnits:    105,
			wantGPU:     11,
			description: "105 MPS units should round up to 11 GPUs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMPSToGPU(tt.mpsUnits)
			assert.Equal(t, tt.wantGPU, got, tt.description)
		})
	}
}

// TestValidateGPUQuota - Comprehensive quota validation
func TestValidateGPUQuota(t *testing.T) {
	tests := []struct {
		name        string
		quota       int
		wantValid   bool
		description string
	}{
		{
			name:        "negative quota",
			quota:       -1,
			wantValid:   false,
			description: "negative quota should be invalid",
		},
		{
			name:        "zero quota",
			quota:       0,
			wantValid:   false,
			description: "zero quota should be invalid",
		},
		{
			name:        "positive quota",
			quota:       1,
			wantValid:   true,
			description: "positive quota should be valid",
		},
		{
			name:        "large quota",
			quota:       10000,
			wantValid:   true,
			description: "large quota should be valid",
		},
		{
			name:        "boundary quota",
			quota:       100,
			wantValid:   true,
			description: "boundary value should be valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateGPUQuota(tt.quota)
			assert.Equal(t, tt.wantValid, got, tt.description)
		})
	}
}

// TestCalculateMPSUsage - Sum of MPS requests
func TestCalculateMPSUsage(t *testing.T) {
	tests := []struct {
		name        string
		requests    []int
		wantTotal   int
		description string
	}{
		{
			name:        "empty requests",
			requests:    []int{},
			wantTotal:   0,
			description: "empty list should sum to 0",
		},
		{
			name:        "single request",
			requests:    []int{10},
			wantTotal:   10,
			description: "single request should return its value",
		},
		{
			name:        "multiple requests",
			requests:    []int{10, 20, 30},
			wantTotal:   60,
			description: "multiple requests should sum correctly",
		},
		{
			name:        "zero in requests",
			requests:    []int{10, 0, 20},
			wantTotal:   30,
			description: "zero values should be included in sum",
		},
		{
			name:        "large request values",
			requests:    []int{1000, 2000, 3000},
			wantTotal:   6000,
			description: "large values should sum correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMPSUsage(tt.requests)
			assert.Equal(t, tt.wantTotal, got, tt.description)
		})
	}
}

// TestIsWithinQuota - Quota boundary checks
func TestIsWithinQuota(t *testing.T) {
	tests := []struct {
		name        string
		requested   int
		available   int
		wantWithin  bool
		description string
	}{
		{
			name:        "zero requested, any available",
			requested:   0,
			available:   100,
			wantWithin:  true,
			description: "zero request should always be within quota",
		},
		{
			name:        "less than available",
			requested:   50,
			available:   100,
			wantWithin:  true,
			description: "requested less than available should be within",
		},
		{
			name:        "equal to available",
			requested:   100,
			available:   100,
			wantWithin:  true,
			description: "requested equal to available should be within",
		},
		{
			name:        "more than available",
			requested:   150,
			available:   100,
			wantWithin:  false,
			description: "requested more than available should be outside",
		},
		{
			name:        "both zero",
			requested:   0,
			available:   0,
			wantWithin:  true,
			description: "zero requested from zero available should be within",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWithinQuota(tt.requested, tt.available)
			assert.Equal(t, tt.wantWithin, got, tt.description)
		})
	}
}

// TestMPSConfigValidate
func TestMPSConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		gpuQuota    int
		memoryMB    int
		wantErr     bool
		description string
	}{
		{
			name:        "valid positive values",
			gpuQuota:    50,
			memoryMB:    2048,
			wantErr:     false,
			description: "positive values should be valid",
		},
		{
			name:        "zero values",
			gpuQuota:    0,
			memoryMB:    0,
			wantErr:     false,
			description: "zero values are allowed",
		},
		{
			name:        "negative GPU quota",
			gpuQuota:    -1,
			memoryMB:    1024,
			wantErr:     false,
			description: "negative GPU quota is allowed (returns nil)",
		},
		{
			name:        "negative memory",
			gpuQuota:    50,
			memoryMB:    -1,
			wantErr:     false,
			description: "negative memory is allowed (returns nil)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MPSConfig{
				GPUQuota:      tt.gpuQuota,
				MemoryLimitMB: tt.memoryMB,
			}
			err := cfg.Validate()

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestMPSConfigToEnvVars - Comprehensive environment variable conversion
func TestMPSConfigToEnvVars(t *testing.T) {
	tests := []struct {
		name           string
		gpuQuota       int
		memoryLimitMB  int
		wantGPUQuota   bool
		wantMemory     bool
		expectedMemory int64
		description    string
	}{
		{
			name:           "both set",
			gpuQuota:       80,
			memoryLimitMB:  2048,
			wantGPUQuota:   true,
			wantMemory:     true,
			expectedMemory: 2048 * 1024 * 1024,
			description:    "both GPU and memory should be set",
		},
		{
			name:          "only GPU quota",
			gpuQuota:      50,
			memoryLimitMB: 0,
			wantGPUQuota:  true,
			wantMemory:    false,
			description:   "only GPU quota should be set",
		},
		{
			name:           "only memory limit",
			gpuQuota:       0,
			memoryLimitMB:  1024,
			wantGPUQuota:   false,
			wantMemory:     true,
			expectedMemory: 1024 * 1024 * 1024,
			description:    "only memory limit should be set",
		},
		{
			name:          "nothing set",
			gpuQuota:      0,
			memoryLimitMB: 0,
			wantGPUQuota:  false,
			wantMemory:    false,
			description:   "no env vars should be set",
		},
		{
			name:          "large GPU quota",
			gpuQuota:      1000,
			memoryLimitMB: 0,
			wantGPUQuota:  true,
			wantMemory:    false,
			description:   "large GPU quota should convert correctly",
		},
		{
			name:           "large memory",
			gpuQuota:       0,
			memoryLimitMB:  8192,
			wantGPUQuota:   false,
			wantMemory:     true,
			expectedMemory: 8192 * 1024 * 1024,
			description:    "large memory should convert to bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &MPSConfig{
				GPUQuota:      tt.gpuQuota,
				MemoryLimitMB: tt.memoryLimitMB,
			}

			env := cfg.ToEnvVars()

			// Check GPU quota
			if tt.wantGPUQuota {
				quotaVal, ok := env["GPU_QUOTA"]
				assert.True(t, ok, tt.description+" - GPU_QUOTA should be set")
				assert.Equal(t, fmt.Sprintf("%d", tt.gpuQuota), quotaVal)
			} else {
				_, ok := env["GPU_QUOTA"]
				assert.False(t, ok, tt.description+" - GPU_QUOTA should not be set")
			}

			// Check memory limit
			if tt.wantMemory {
				memVal, ok := env["CUDA_MPS_PINNED_DEVICE_MEM_LIMIT"]
				assert.True(t, ok, tt.description+" - CUDA_MPS_PINNED_DEVICE_MEM_LIMIT should be set")
				assert.Equal(t, fmt.Sprintf("%d", tt.expectedMemory), memVal)
			} else {
				_, ok := env["CUDA_MPS_PINNED_DEVICE_MEM_LIMIT"]
				assert.False(t, ok, tt.description+" - CUDA_MPS_PINNED_DEVICE_MEM_LIMIT should not be set")
			}
		})
	}
}

// TestProjectMPSQuota - Comprehensive quota management
func TestProjectMPSQuota(t *testing.T) {
	t.Run("AvailableMPS calculation", func(t *testing.T) {
		tests := []struct {
			name        string
			total       int
			used        int
			wantAvail   int
			description string
		}{
			{
				name:        "no usage",
				total:       100,
				used:        0,
				wantAvail:   100,
				description: "available should equal total when nothing used",
			},
			{
				name:        "partial usage",
				total:       100,
				used:        40,
				wantAvail:   60,
				description: "available should be total minus used",
			},
			{
				name:        "fully used",
				total:       100,
				used:        100,
				wantAvail:   0,
				description: "available should be zero when fully used",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := &ProjectMPSQuota{
					ProjectID:     1,
					TotalMPSUnits: tt.total,
					UsedMPSUnits:  tt.used,
				}
				assert.Equal(t, tt.wantAvail, q.AvailableMPS(), tt.description)
			})
		}
	})

	t.Run("CanAllocate checks", func(t *testing.T) {
		tests := []struct {
			name        string
			total       int
			used        int
			request     int
			wantAllow   bool
			description string
		}{
			{
				name:        "allocate within available",
				total:       100,
				used:        40,
				request:     50,
				wantAllow:   true,
				description: "should not allocate when not enough available",
			},
			{
				name:        "allocate equal to available",
				total:       100,
				used:        40,
				request:     60,
				wantAllow:   true,
				description: "should allocate when request equals available",
			},
			{
				name:        "allocate more than available",
				total:       100,
				used:        40,
				request:     70,
				wantAllow:   false,
				description: "should not allocate more than available",
			},
			{
				name:        "zero request",
				total:       100,
				used:        50,
				request:     0,
				wantAllow:   true,
				description: "zero request should always be allowed",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := &ProjectMPSQuota{
					ProjectID:     1,
					TotalMPSUnits: tt.total,
					UsedMPSUnits:  tt.used,
				}
				assert.Equal(t, tt.wantAllow, q.CanAllocate(tt.request), tt.description)
			})
		}
	})

	t.Run("Allocate functionality", func(t *testing.T) {
		tests := []struct {
			name        string
			total       int
			initial     int
			request     int
			wantErr     bool
			wantUsed    int
			description string
		}{
			{
				name:        "successful allocation",
				total:       100,
				initial:     40,
				request:     30,
				wantErr:     false,
				wantUsed:    70,
				description: "should increase used after allocation",
			},
			{
				name:        "negative request",
				total:       100,
				initial:     40,
				request:     -10,
				wantErr:     false,
				wantUsed:    40,
				description: "negative request should not affect used",
			},
			{
				name:        "insufficient quota",
				total:       100,
				initial:     90,
				request:     20,
				wantErr:     false,
				wantUsed:    90,
				description: "should not allocate when insufficient quota",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := &ProjectMPSQuota{
					ProjectID:     1,
					TotalMPSUnits: tt.total,
					UsedMPSUnits:  tt.initial,
				}
				err := q.Allocate(tt.request)

				if tt.wantErr {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
					assert.Equal(t, tt.wantUsed, q.UsedMPSUnits, tt.description)
				}
			})
		}
	})

	t.Run("Release functionality", func(t *testing.T) {
		tests := []struct {
			name        string
			total       int
			used        int
			release     int
			wantErr     bool
			wantUsed    int
			description string
		}{
			{
				name:        "release partial",
				total:       100,
				used:        50,
				release:     20,
				wantErr:     false,
				wantUsed:    30,
				description: "should decrease used after release",
			},
			{
				name:        "release all",
				total:       100,
				used:        50,
				release:     50,
				wantErr:     false,
				wantUsed:    0,
				description: "should allow releasing all used units",
			},
			{
				name:        "release more than used",
				total:       100,
				used:        30,
				release:     50,
				wantErr:     false,
				wantUsed:    0,
				description: "should clamp to zero when release exceeds used",
			},
			{
				name:        "negative release",
				total:       100,
				used:        50,
				release:     -10,
				wantErr:     false,
				wantUsed:    50,
				description: "negative release should not affect used",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := &ProjectMPSQuota{
					ProjectID:     1,
					TotalMPSUnits: tt.total,
					UsedMPSUnits:  tt.used,
				}
				err := q.Release(tt.release)

				if tt.wantErr {
					assert.Error(t, err, tt.description)
				} else {
					assert.NoError(t, err, tt.description)
					assert.Equal(t, tt.wantUsed, q.UsedMPSUnits, tt.description)
				}
			})
		}
	})

	t.Run("UsagePercent calculation", func(t *testing.T) {
		tests := []struct {
			name        string
			total       int
			used        int
			wantPercent float64
			description string
		}{
			{
				name:        "zero total",
				total:       0,
				used:        0,
				wantPercent: 0,
				description: "zero total should return 0%",
			},
			{
				name:        "no usage",
				total:       100,
				used:        0,
				wantPercent: 0,
				description: "no usage should be 0%",
			},
			{
				name:        "partial usage",
				total:       100,
				used:        40,
				wantPercent: 40,
				description: "partial usage should calculate correctly",
			},
			{
				name:        "full usage",
				total:       100,
				used:        100,
				wantPercent: 100,
				description: "full usage should be 100%",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				q := &ProjectMPSQuota{
					ProjectID:     1,
					TotalMPSUnits: tt.total,
					UsedMPSUnits:  tt.used,
				}
				got := q.UsagePercent()
				assert.InDelta(t, tt.wantPercent, got, 0.01, tt.description)
			})
		}
	})
}
