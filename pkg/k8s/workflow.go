package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var workflowGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "workflows",
}

func DeleteWorkflow(ctx context.Context, namespace, name string) error {
	if DynamicClient == nil {
		slog.Debug("[MOCK] delete workflow", slog.String("namespace", namespace), slog.String("name", name))
		return nil
	}
	if namespace == "" {
		namespace = "default"
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	policy := metav1.DeletePropagationBackground
	err := DynamicClient.Resource(workflowGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &policy})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete workflow: %w", err)
	}
	return nil
}

func GetWorkflowPhase(ctx context.Context, namespace, name string) (string, error) {
	if DynamicClient == nil {
		slog.Debug("[MOCK] get workflow phase", slog.String("namespace", namespace), slog.String("name", name))
		return "", nil
	}
	if namespace == "" {
		namespace = "default"
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	obj, err := DynamicClient.Resource(workflowGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("get workflow: %w", err)
	}
	if obj == nil {
		return "", nil
	}
	status, ok := obj.Object["status"].(map[string]interface{})
	if !ok || status == nil {
		return "", nil
	}
	phaseRaw, ok := status["phase"].(string)
	if !ok {
		return "", nil
	}
	return strings.TrimSpace(phaseRaw), nil
}
