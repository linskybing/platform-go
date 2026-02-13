package cluster

import "time"

type GPUDeviceInfo struct {
	Index       int     `json:"index"`
	UUID        string  `json:"uuid,omitempty"`
	Utilization float64 `json:"utilization"`
	MemoryBytes int64   `json:"memoryBytes"`
}

type NodeResourceInfo struct {
	Name                   string          `json:"name"`
	CPUAllocatableMilli    int64           `json:"cpuAllocatableMilli"`
	CPUUsedMilli           int64           `json:"cpuUsedMilli"`
	MemoryAllocatableBytes int64           `json:"memoryAllocatableBytes"`
	MemoryUsedBytes        int64           `json:"memoryUsedBytes"`
	GPUAllocatable         int64           `json:"gpuAllocatable"`
	GPUUsed                int64           `json:"gpuUsed"`
	GPUDevices             []GPUDeviceInfo `json:"gpuDevices,omitempty"`
}

type PodGPUUsage struct {
	PodName     string  `json:"podName"`
	Namespace   string  `json:"namespace"`
	Node        string  `json:"node,omitempty"`
	GPUIndex    int     `json:"gpuIndex"`
	GPUUUID     string  `json:"gpuUuid,omitempty"`
	MemoryBytes int64   `json:"memoryBytes"`
	Utilization float64 `json:"utilization"`
}

type ClusterSummary struct {
	NodeCount                   int                `json:"nodeCount"`
	TotalCPUAllocatableMilli    int64              `json:"totalCpuAllocatableMilli"`
	TotalCPUUsedMilli           int64              `json:"totalCpuUsedMilli"`
	TotalMemoryAllocatableBytes int64              `json:"totalMemoryAllocatableBytes"`
	TotalMemoryUsedBytes        int64              `json:"totalMemoryUsedBytes"`
	TotalGPUAllocatable         int64              `json:"totalGpuAllocatable"`
	TotalGPUUsed                int64              `json:"totalGpuUsed"`
	Nodes                       []NodeResourceInfo `json:"nodes"`
	PodGPUUsages                []PodGPUUsage      `json:"podGpuUsages,omitempty"`
	CollectedAt                 time.Time          `json:"collectedAt"`
}
