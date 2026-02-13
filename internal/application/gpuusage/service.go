package gpuusage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/linskybing/platform-go/internal/application/executor"
	domain "github.com/linskybing/platform-go/internal/domain/gpuusage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/gpu"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/prometheus"
	"github.com/prometheus/common/model"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	jobLabelKey = "platform.job-id"
)

type GPUUsageService struct {
	repos *repository.Repos
	prom  *prometheus.Client
}

func NewGPUUsageService(repos *repository.Repos, promClient *prometheus.Client) *GPUUsageService {
	return &GPUUsageService{repos: repos, prom: promClient}
}

func (s *GPUUsageService) CollectSnapshots(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.prom == nil {
		return fmt.Errorf("prometheus client not configured")
	}
	if k8s.Clientset == nil {
		return fmt.Errorf("kubernetes client not configured")
	}

	jobs, err := s.repos.Job.ListByStatus(ctx, []string{string(executor.JobStatusRunning)})
	if err != nil {
		return fmt.Errorf("list running jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil
	}

	var snapshots []domain.JobGPUUsageSnapshot
	for _, job := range jobs {
		pods, err := s.listJobPods(ctx, job.Namespace, job.ID)
		if err != nil {
			slog.Warn("list job pods failed", "job_id", job.ID, "error", err)
			continue
		}
		if len(pods) == 0 {
			slog.Warn("no pods found for job labels", "job_id", job.ID, "namespace", job.Namespace)
			continue
		}
		for _, pod := range pods {
			podSnapshots, err := s.collectPodSnapshots(ctx, job.ID, &pod)
			if err != nil {
				slog.Warn("collect pod snapshots failed", "pod", pod.Name, "job_id", job.ID, "error", err)
				continue
			}
			snapshots = append(snapshots, podSnapshots...)
		}
	}

	return s.repos.GPUUsage.InsertSnapshots(ctx, snapshots)
}

func (s *GPUUsageService) ListSnapshots(ctx context.Context, jobID string, limit, offset int) ([]domain.JobGPUUsageSnapshot, int64, error) {
	return s.repos.GPUUsage.ListSnapshotsByJob(ctx, jobID, limit, offset)
}

func (s *GPUUsageService) GetSummary(ctx context.Context, jobID string) (*domain.JobGPUUsageSummary, error) {
	return s.repos.GPUUsage.GetSummary(ctx, jobID)
}

func (s *GPUUsageService) ComputeMissingSummaries(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	statuses := []string{string(executor.JobStatusCompleted), string(executor.JobStatusFailed), string(executor.JobStatusCancelled)}
	jobs, err := s.repos.Job.ListByStatus(ctx, statuses)
	if err != nil {
		return fmt.Errorf("list terminal jobs: %w", err)
	}

	for _, job := range jobs {
		_, err := s.repos.GPUUsage.GetSummary(ctx, job.ID)
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("get summary: %w", err)
		}

		summary, err := s.computeSummary(ctx, job.ID)
		if err != nil {
			slog.Warn("compute gpu summary failed", "job_id", job.ID, "error", err)
			continue
		}
		if summary != nil {
			if err := s.repos.GPUUsage.UpsertSummary(ctx, summary); err != nil {
				return fmt.Errorf("upsert summary: %w", err)
			}
		}
	}
	return nil
}

func (s *GPUUsageService) CleanupSnapshots(ctx context.Context, maxAge time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cutoff := time.Now().Add(-maxAge)
	return s.repos.GPUUsage.DeleteSnapshotsBefore(ctx, cutoff)
}

func (s *GPUUsageService) listJobPods(ctx context.Context, namespace, jobID string) ([]corev1.Pod, error) {
	if namespace == "" || jobID == "" {
		return nil, nil
	}

	selector := fmt.Sprintf("%s=%s", jobLabelKey, jobID)
	pods, err := k8s.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}
	if pods != nil && len(pods.Items) > 0 {
		return pods.Items, nil
	}
	return nil, nil
}

func (s *GPUUsageService) collectPodSnapshots(ctx context.Context, jobID string, pod *corev1.Pod) ([]domain.JobGPUUsageSnapshot, error) {
	if pod == nil {
		return nil, nil
	}
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return nil, nil
	}

	timestamp := time.Now()
	memoryVector, memErr := s.queryPodMetric(ctx, "flashsched_pod_gpu_mps_memory_bytes", pod.Namespace, pod.Name)
	if memErr != nil {
		slog.Warn("pod gpu memory query failed", "pod", pod.Name, "namespace", pod.Namespace, "error", memErr)
	}
	utilVector, utilErr := s.queryPodMetric(ctx, "flashsched_pod_gpu_sm_utilization_ratio", pod.Namespace, pod.Name)
	if utilErr != nil {
		slog.Warn("pod gpu utilization query failed", "pod", pod.Name, "namespace", pod.Namespace, "error", utilErr)
	}
	if memErr != nil && utilErr != nil {
		return nil, fmt.Errorf("both pod gpu metrics unavailable")
	}

	mpsUnits := podMPSUnits(pod)
	usageMap := make(map[string]*domain.JobGPUUsageSnapshot)
	mergeSample := func(sample *model.Sample, isUtil bool) {
		idx, idxOK := parseGPUIndex(sample)
		uuid := string(sample.Metric[model.LabelName("gpu_uuid")])
		node := string(sample.Metric[model.LabelName("node")])
		if node == "" {
			node = pod.Spec.NodeName
		}

		key := fmt.Sprintf("%s/%s:%d:%s", pod.Namespace, pod.Name, idx, uuid)
		entry := usageMap[key]
		if entry == nil {
			entry = &domain.JobGPUUsageSnapshot{
				JobID:               jobID,
				Timestamp:           timestamp,
				PodName:             pod.Name,
				PodNamespace:        pod.Namespace,
				GPUIndex:            idx,
				GPUUUID:             uuid,
				Node:                node,
				MPSVirtualUnits:     mpsUnits,
				MPSPhysicalGPUIndex: idx,
			}
			if !idxOK {
				entry.MPSPhysicalGPUIndex = -1
			}
			usageMap[key] = entry
		}

		if isUtil {
			entry.GPUUtilization += float64(sample.Value)
			return
		}
		entry.GPUMemoryBytes += int64(sample.Value)
	}

	for _, sample := range memoryVector {
		mergeSample(sample, false)
	}
	for _, sample := range utilVector {
		mergeSample(sample, true)
	}
	if len(usageMap) == 0 {
		return nil, nil
	}

	result := make([]domain.JobGPUUsageSnapshot, 0, len(usageMap))
	for _, entry := range usageMap {
		result = append(result, *entry)
	}
	return result, nil
}

func (s *GPUUsageService) queryPodMetric(ctx context.Context, metric, namespace, pod string) (model.Vector, error) {
	if s.prom == nil {
		return nil, fmt.Errorf("prometheus client not configured")
	}
	query := fmt.Sprintf(`%s{pod_namespace="%s",pod_name="%s"}`, metric, namespace, pod)
	val, _, err := s.prom.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	vector, ok := val.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected prometheus result type")
	}
	return vector, nil
}

func parseGPUIndex(sample *model.Sample) (int, bool) {
	idxRaw := string(sample.Metric[model.LabelName("gpu_index")])
	if idxRaw == "" {
		return -1, false
	}
	idx, err := strconv.Atoi(idxRaw)
	if err != nil {
		return -1, false
	}
	return idx, true
}

func podMPSUnits(pod *corev1.Pod) int {
	if pod == nil {
		return 0
	}
	totalGPU := int64(0)
	for _, container := range pod.Spec.Containers {
		if qty, ok := container.Resources.Requests[corev1.ResourceName("nvidia.com/gpu")]; ok {
			totalGPU += qty.Value()
		}
	}
	if totalGPU <= 0 {
		return 0
	}
	return gpu.ConvertGPUToMPS(int(totalGPU))
}

func (s *GPUUsageService) computeSummary(ctx context.Context, jobID string) (*domain.JobGPUUsageSummary, error) {
	snapshots, err := s.repos.GPUUsage.ListAllSnapshotsByJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, nil
	}

	var sumUtil float64
	var sumMem int64
	var peakMem int64
	first := snapshots[0].Timestamp
	last := snapshots[0].Timestamp

	timeSlots := make(map[time.Time]map[string]struct{})
	for _, snap := range snapshots {
		if snap.Timestamp.Before(first) {
			first = snap.Timestamp
		}
		if snap.Timestamp.After(last) {
			last = snap.Timestamp
		}
		if snap.GPUMemoryBytes > peakMem {
			peakMem = snap.GPUMemoryBytes
		}
		sumUtil += snap.GPUUtilization
		sumMem += snap.GPUMemoryBytes

		bucket := snap.Timestamp.Truncate(time.Second)
		key := fmt.Sprintf("%s/%s:%d:%s", snap.PodNamespace, snap.PodName, snap.GPUIndex, snap.GPUUUID)
		set := timeSlots[bucket]
		if set == nil {
			set = make(map[string]struct{})
			timeSlots[bucket] = set
		}
		set[key] = struct{}{}
	}

	buckets := make([]time.Time, 0, len(timeSlots))
	for ts := range timeSlots {
		buckets = append(buckets, ts)
	}
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].Before(buckets[j]) })
	interval := medianInterval(buckets)

	var totalGPUSeconds float64
	for i := 0; i < len(buckets); i++ {
		dt := interval
		if i+1 < len(buckets) {
			next := buckets[i+1].Sub(buckets[i])
			if next > 0 {
				dt = next
			}
		}
		gpuCount := float64(len(timeSlots[buckets[i]]))
		totalGPUSeconds += dt.Seconds() * gpuCount
	}

	sampleCount := len(snapshots)
	avgUtil := sumUtil / float64(sampleCount)
	avgMem := int64(0)
	if sampleCount > 0 {
		avgMem = sumMem / int64(sampleCount)
	}

	summary := &domain.JobGPUUsageSummary{
		JobID:           jobID,
		TotalGPUSeconds: totalGPUSeconds,
		PeakMemoryBytes: peakMem,
		AvgUtilization:  avgUtil,
		AvgMemoryBytes:  avgMem,
		SampleCount:     sampleCount,
		FirstSampleAt:   &first,
		LastSampleAt:    &last,
		ComputedAt:      time.Now(),
	}
	return summary, nil
}

func medianInterval(buckets []time.Time) time.Duration {
	if len(buckets) < 2 {
		return 30 * time.Second
	}

	diffs := make([]time.Duration, 0, len(buckets)-1)
	for i := 0; i < len(buckets)-1; i++ {
		delta := buckets[i+1].Sub(buckets[i])
		if delta > 0 {
			diffs = append(diffs, delta)
		}
	}
	if len(diffs) == 0 {
		return 30 * time.Second
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i] < diffs[j] })
	mid := len(diffs) / 2
	if len(diffs)%2 == 1 {
		return diffs[mid]
	}
	return (diffs[mid-1] + diffs[mid]) / 2
}
