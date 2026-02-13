package gpuusage

import "time"

type JobGPUUsageSnapshot struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement"`
	JobID               string    `gorm:"size:21;not null;index:idx_job_time,priority:1"`
	Timestamp           time.Time `gorm:"not null;index:idx_job_time,priority:2;index:idx_snapshot_time"`
	PodName             string    `gorm:"size:255;not null;index:idx_job_pod_time,priority:2"`
	PodNamespace        string    `gorm:"size:255;not null;index:idx_job_pod_time,priority:1"`
	GPUIndex            int       `gorm:"not null"`
	GPUUUID             string    `gorm:"size:100"`
	GPUMemoryBytes      int64     `gorm:"default:0"`
	GPUUtilization      float64   `gorm:"default:0"`
	Node                string    `gorm:"size:255;not null"`
	MPSVirtualUnits     int       `gorm:"default:0"`
	MPSPhysicalGPUIndex int       `gorm:"default:-1"`
}

func (JobGPUUsageSnapshot) TableName() string {
	return "job_gpu_usage_snapshots"
}

type JobGPUUsageSummary struct {
	ID              int64   `gorm:"primaryKey;autoIncrement"`
	JobID           string  `gorm:"size:21;not null;uniqueIndex"`
	TotalGPUSeconds float64 `gorm:"default:0"`
	PeakMemoryBytes int64   `gorm:"default:0"`
	AvgUtilization  float64 `gorm:"default:0"`
	AvgMemoryBytes  int64   `gorm:"default:0"`
	SampleCount     int     `gorm:"default:0"`
	FirstSampleAt   *time.Time
	LastSampleAt    *time.Time
	ComputedAt      time.Time `gorm:"autoCreateTime"`
}

func (JobGPUUsageSummary) TableName() string {
	return "job_gpu_usage_summaries"
}
