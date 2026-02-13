package cluster

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	domain "github.com/linskybing/platform-go/internal/domain/cluster"
	"github.com/linskybing/platform-go/pkg/cache"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/prometheus"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterCacheKey = "cluster:summary"
	// DefaultClusterCacheTTL is the default TTL for cluster summary cache entries.
	DefaultClusterCacheTTL = 30 * time.Second
)

type ClusterService struct {
	cache *cache.Service
	prom  *prometheus.Client
	ttl   time.Duration
}

func NewClusterService(cacheSvc *cache.Service, promClient *prometheus.Client) *ClusterService {
	return &ClusterService{
		cache: cacheSvc,
		prom:  promClient,
		ttl:   DefaultClusterCacheTTL,
	}
}

func (s *ClusterService) GetSummary(ctx context.Context) (domain.ClusterSummary, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var summary domain.ClusterSummary
	if s.cache == nil || !s.cache.Enabled() {
		return s.CollectClusterResources(ctx)
	}

	err := s.cache.GetOrFetchJSON(ctx, clusterCacheKey, s.ttl, &summary, func(ctx context.Context) (interface{}, error) {
		return s.CollectClusterResources(ctx)
	})
	return summary, err
}

func (s *ClusterService) RefreshCache(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	summary, err := s.CollectClusterResources(ctx)
	if err != nil {
		return err
	}
	if s.cache == nil || !s.cache.Enabled() {
		return nil
	}
	return s.cache.SetJSON(ctx, clusterCacheKey, summary, s.ttl)
}

func (s *ClusterService) CollectClusterResources(ctx context.Context) (domain.ClusterSummary, error) {
	if k8s.Clientset == nil {
		return domain.ClusterSummary{}, fmt.Errorf("kubernetes client not configured")
	}

	nodes, err := k8s.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return domain.ClusterSummary{}, fmt.Errorf("list nodes: %w", err)
	}
	pods, err := k8s.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return domain.ClusterSummary{}, fmt.Errorf("list pods: %w", err)
	}

	nodeMap := make(map[string]*domain.NodeResourceInfo, len(nodes.Items))
	summary := domain.ClusterSummary{CollectedAt: time.Now()}
	for _, node := range nodes.Items {
		allocCPU := quantityMilli(node.Status.Allocatable, corev1.ResourceCPU)
		allocMem := quantityValue(node.Status.Allocatable, corev1.ResourceMemory)
		allocGPU := quantityValue(node.Status.Allocatable, corev1.ResourceName("nvidia.com/gpu"))

		info := &domain.NodeResourceInfo{
			Name:                   node.Name,
			CPUAllocatableMilli:    allocCPU,
			MemoryAllocatableBytes: allocMem,
			GPUAllocatable:         allocGPU,
		}
		nodeMap[node.Name] = info
		summary.TotalCPUAllocatableMilli += allocCPU
		summary.TotalMemoryAllocatableBytes += allocMem
		summary.TotalGPUAllocatable += allocGPU
	}

	for _, pod := range pods.Items {
		if pod.Spec.NodeName == "" {
			continue
		}
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		info := nodeMap[pod.Spec.NodeName]
		if info == nil {
			continue
		}

		for _, container := range pod.Spec.Containers {
			cpu := containerQuantityMilli(container.Resources, corev1.ResourceCPU)
			mem := containerQuantityValue(container.Resources, corev1.ResourceMemory)
			gpu := containerQuantityValue(container.Resources, corev1.ResourceName("nvidia.com/gpu"))

			info.CPUUsedMilli += cpu
			info.MemoryUsedBytes += mem
			info.GPUUsed += gpu
			summary.TotalCPUUsedMilli += cpu
			summary.TotalMemoryUsedBytes += mem
			summary.TotalGPUUsed += gpu
		}
	}

	gpuMetrics := s.queryGPUUtilization(ctx)
	podUsages := s.queryPodGPUUsage(ctx)

	for nodeName, gpus := range gpuMetrics {
		info := nodeMap[nodeName]
		if info == nil {
			continue
		}
		indices := make([]int, 0, len(gpus))
		for idx := range gpus {
			indices = append(indices, idx)
		}
		sort.Ints(indices)
		for _, idx := range indices {
			metric := gpus[idx]
			info.GPUDevices = append(info.GPUDevices, domain.GPUDeviceInfo{
				Index:       idx,
				UUID:        metric.uuid,
				Utilization: metric.utilization,
			})
		}
	}

	summary.Nodes = make([]domain.NodeResourceInfo, 0, len(nodeMap))
	for _, info := range nodeMap {
		summary.Nodes = append(summary.Nodes, *info)
	}
	sort.Slice(summary.Nodes, func(i, j int) bool {
		return summary.Nodes[i].Name < summary.Nodes[j].Name
	})
	if len(podUsages) > 0 {
		summary.PodGPUUsages = podUsages
	}
	summary.NodeCount = len(summary.Nodes)

	return summary, nil
}

type gpuMetric struct {
	utilization float64
	uuid        string
}

func (s *ClusterService) queryGPUUtilization(ctx context.Context) map[string]map[int]gpuMetric {
	result := make(map[string]map[int]gpuMetric)
	if s.prom == nil {
		return result
	}

	val, _, err := s.prom.Query(ctx, "flashsched_node_gpu_mps_utilization_ratio{node=~\".*\"}", time.Now())
	if err != nil {
		slog.Warn("prometheus GPU utilization query failed", "error", err)
		return result
	}

	vector, ok := val.(model.Vector)
	if !ok {
		return result
	}

	for _, sample := range vector {
		node := string(sample.Metric[model.LabelName("node")])
		idxRaw := string(sample.Metric[model.LabelName("gpu_index")])
		if node == "" || idxRaw == "" {
			continue
		}
		idx, err := strconv.Atoi(idxRaw)
		if err != nil {
			continue
		}
		uuid := string(sample.Metric[model.LabelName("gpu_uuid")])
		util := float64(sample.Value)

		if _, ok := result[node]; !ok {
			result[node] = make(map[int]gpuMetric)
		}
		result[node][idx] = gpuMetric{utilization: util, uuid: uuid}
	}

	return result
}

func (s *ClusterService) queryPodGPUUsage(ctx context.Context) []domain.PodGPUUsage {
	if s.prom == nil {
		return nil
	}

	memoryVal, _, memErr := s.prom.Query(ctx, "flashsched_pod_gpu_mps_memory_bytes{pod_namespace=~\".*\"}", time.Now())
	if memErr != nil {
		slog.Warn("prometheus pod GPU memory query failed", "error", memErr)
		return nil
	}
	utilVal, _, utilErr := s.prom.Query(ctx, "flashsched_pod_gpu_sm_utilization_ratio{pod_namespace=~\".*\"}", time.Now())
	if utilErr != nil {
		slog.Warn("prometheus pod GPU utilization query failed", "error", utilErr)
		return nil
	}

	memoryVector, ok := memoryVal.(model.Vector)
	if !ok {
		return nil
	}
	utilVector, ok := utilVal.(model.Vector)
	if !ok {
		return nil
	}

	usageMap := make(map[string]*domain.PodGPUUsage, len(memoryVector))
	mergeSample := func(sample *model.Sample, isUtil bool) {
		pod := string(sample.Metric[model.LabelName("pod_name")])
		ns := string(sample.Metric[model.LabelName("pod_namespace")])
		if pod == "" || ns == "" {
			return
		}
		node := string(sample.Metric[model.LabelName("node")])
		idxRaw := string(sample.Metric[model.LabelName("gpu_index")])
		idx := -1
		if idxRaw != "" {
			if parsed, err := strconv.Atoi(idxRaw); err == nil {
				idx = parsed
			}
		}
		uuid := string(sample.Metric[model.LabelName("gpu_uuid")])

		key := fmt.Sprintf("%s/%s:%s:%d:%s", ns, pod, node, idx, uuid)
		usage := usageMap[key]
		if usage == nil {
			usage = &domain.PodGPUUsage{
				PodName:   pod,
				Namespace: ns,
				Node:      node,
				GPUIndex:  idx,
				GPUUUID:   uuid,
			}
			usageMap[key] = usage
		}
		if isUtil {
			usage.Utilization += float64(sample.Value)
			return
		}
		usage.MemoryBytes += int64(sample.Value)
	}

	for _, sample := range memoryVector {
		mergeSample(sample, false)
	}
	for _, sample := range utilVector {
		mergeSample(sample, true)
	}

	usages := make([]domain.PodGPUUsage, 0, len(usageMap))
	for _, usage := range usageMap {
		usages = append(usages, *usage)
	}
	return usages
}

func containerQuantityMilli(resources corev1.ResourceRequirements, name corev1.ResourceName) int64 {
	if qty, ok := resources.Requests[name]; ok {
		return qty.MilliValue()
	}
	if qty, ok := resources.Limits[name]; ok {
		return qty.MilliValue()
	}
	return 0
}

func containerQuantityValue(resources corev1.ResourceRequirements, name corev1.ResourceName) int64 {
	if qty, ok := resources.Requests[name]; ok {
		return qty.Value()
	}
	if qty, ok := resources.Limits[name]; ok {
		return qty.Value()
	}
	return 0
}

func quantityMilli(list corev1.ResourceList, name corev1.ResourceName) int64 {
	qty, ok := list[name]
	if !ok {
		return 0
	}
	return qty.MilliValue()
}

func quantityValue(list corev1.ResourceList, name corev1.ResourceName) int64 {
	qty, ok := list[name]
	if !ok {
		return 0
	}
	return qty.Value()
}
