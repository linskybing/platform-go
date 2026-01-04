package mps

import (
	"fmt"
	"testing"
)

func TestConvertGPUToMPS(t *testing.T) {
	if got := ConvertGPUToMPS(2); got != 20 {
		t.Fatalf("expected 20 MPS units, got %d", got)
	}
}

func TestConvertMPSToGPU(t *testing.T) {
	if got := ConvertMPSToGPU(15); got != 2 {
		t.Fatalf("expected 2 GPUs for 15 units, got %d", got)
	}
}

func TestValidateMPSLimit(t *testing.T) {
	if !ValidateMPSLimit(100) {
		t.Fatalf("expected limit 100 to be valid")
	}
	if ValidateMPSLimit(-1) {
		t.Fatalf("expected negative limit to be invalid")
	}
}

func TestProjectMPSQuota(t *testing.T) {
	q := &ProjectMPSQuota{TotalMPSUnits: 100, UsedMPSUnits: 40}
	if !q.CanAllocate(50) {
		t.Fatalf("expected to allocate 50 units")
	}
	if q.CanAllocate(70) {
		t.Fatalf("expected 70 allocation to fail")
	}
	if q.UsagePercent() != 40 {
		t.Fatalf("expected usage percent 40, got %f", q.UsagePercent())
	}
}

// TestMPSConfigToEnvVars verifies that MPS configuration is correctly converted to CUDA environment variables
func TestMPSConfigToEnvVars(t *testing.T) {
	t.Run("ThreadPercentage and MemoryLimit both set", func(t *testing.T) {
		cfg := &MPSConfig{
			ThreadPercentage: 80,
			MemoryLimitMB:    2048,
		}

		env := cfg.ToEnvVars()

		// Check thread percentage env var
		if threadVal, ok := env["CUDA_MPS_ACTIVE_THREAD_PERCENTAGE"]; !ok {
			t.Fatalf("expected CUDA_MPS_ACTIVE_THREAD_PERCENTAGE to be set")
		} else if threadVal != "80" {
			t.Fatalf("expected thread percentage 80, got %s", threadVal)
		}

		// Check memory limit env var (should be in bytes)
		if memVal, ok := env["CUDA_MPS_PINNED_DEVICE_MEM_LIMIT"]; !ok {
			t.Fatalf("expected CUDA_MPS_PINNED_DEVICE_MEM_LIMIT to be set")
		} else {
			expectedBytes := int64(2048) * 1024 * 1024
			expectedStr := fmt.Sprintf("%d", expectedBytes)
			if memVal != expectedStr {
				t.Fatalf("expected memory limit %s bytes, got %s", expectedStr, memVal)
			}
		}
	})

	t.Run("Only MemoryLimit set", func(t *testing.T) {
		cfg := &MPSConfig{
			ThreadPercentage: 0,
			MemoryLimitMB:    1024,
		}

		env := cfg.ToEnvVars()

		// Should not have thread percentage env var
		if _, ok := env["CUDA_MPS_ACTIVE_THREAD_PERCENTAGE"]; ok {
			t.Fatalf("should not set thread percentage when 0")
		}

		// Should have memory limit env var
		if _, ok := env["CUDA_MPS_PINNED_DEVICE_MEM_LIMIT"]; !ok {
			t.Fatalf("expected CUDA_MPS_PINNED_DEVICE_MEM_LIMIT to be set")
		}
	})

	t.Run("No configuration set", func(t *testing.T) {
		cfg := &MPSConfig{
			ThreadPercentage: 0,
			MemoryLimitMB:    0,
		}

		env := cfg.ToEnvVars()

		// Should be empty
		if len(env) != 0 {
			t.Fatalf("expected empty env vars, got %v", env)
		}
	})
}
