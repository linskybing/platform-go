package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var queueGVR = schema.GroupVersionResource{
	Group:    "scheduling.flash-sched.io",
	Version:  "v1alpha1",
	Resource: "clusterresourcequeues",
}

func EnsureConfigFileQueue(ctx context.Context) error {
	if DynamicClient == nil {
		slog.Debug("[MOCK] ensure cluster resource queues")
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	if config.ConfigFileQueueName != "" {
		if err := ensureQueue(
			ctx,
			config.ConfigFileQueueName,
			config.ConfigFileQueuePriority,
			config.ConfigFileQueuePreemptible,
			config.ConfigFileQueueMaxConcurrent,
			config.ConfigFileQueueTTLSeconds,
		); err != nil {
			return err
		}
	}
	if config.ConfigFileJobQueueName != "" {
		if err := ensureQueue(
			ctx,
			config.ConfigFileJobQueueName,
			config.ConfigFileJobQueuePriority,
			config.ConfigFileJobQueuePreemptible,
			config.ConfigFileJobQueueMaxConcurrent,
			config.ConfigFileJobQueueTTLSeconds,
		); err != nil {
			return err
		}
	}
	if config.DefaultQueueName != "" {
		if err := ensureQueue(
			ctx,
			config.DefaultQueueName,
			config.DefaultQueuePriority,
			config.DefaultQueuePreemptible,
			config.DefaultQueueMaxConcurrent,
			config.DefaultQueueTTLSeconds,
		); err != nil {
			return err
		}
	}
	return nil
}

func ensureQueue(ctx context.Context, name string, priority int64, preemptible bool, maxConcurrent int64, ttlSeconds int64) error {
	if name == "" {
		return nil
	}
	_, err := DynamicClient.Resource(queueGVR).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("get queue %s: %w", name, err)
	}

	crq := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "scheduling.flash-sched.io/v1alpha1",
			"kind":       "ClusterResourceQueue",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"priority":      priority,
				"isPreemptible": preemptible,
				"maxConcurrent": maxConcurrent,
				"ttlSeconds":    ttlSeconds,
			},
		},
	}

	_, err = DynamicClient.Resource(queueGVR).Create(ctx, crq, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create queue %s: %w", name, err)
	}
	return nil
}
