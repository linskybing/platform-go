package k8stest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// TestCluster holds K8s test cluster configuration
type TestCluster struct {
	Client    *kubernetes.Clientset
	Namespace string
	Cleanup   []func() error
}

// SetupTestCluster initializes connection to KinD cluster for testing
func SetupTestCluster(t *testing.T) *TestCluster {
	t.Helper()

	// Skip if not running integration tests
	if os.Getenv("SKIP_K8S_TESTS") == "true" {
		t.Skip("Skipping K8s integration tests (SKIP_K8S_TESTS=true)")
	}

	// Load kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get home directory: %v", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Build config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("failed to build kubeconfig: %v", err)
	}

	// Create clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("failed to create K8s client: %v", err)
	}

	// Create test namespace
	namespace := fmt.Sprintf("test-%s-%d", t.Name(), time.Now().Unix())
	namespace = sanitizeNamespaceName(namespace)

	ctx := context.Background()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"test":       "true",
				"created-by": "integration-test",
			},
		},
	}

	_, err = client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create test namespace: %v", err)
	}

	tc := &TestCluster{
		Client:    client,
		Namespace: namespace,
		Cleanup:   []func() error{},
	}

	// Register cleanup
	t.Cleanup(func() {
		tc.TearDown(t)
	})

	return tc
}

// TearDown cleans up test resources
func (tc *TestCluster) TearDown(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Run custom cleanup functions
	for _, cleanup := range tc.Cleanup {
		if err := cleanup(); err != nil {
			t.Logf("cleanup error: %v", err)
		}
	}

	// Delete test namespace
	if tc.Namespace != "" {
		err := tc.Client.CoreV1().Namespaces().Delete(ctx, tc.Namespace, metav1.DeleteOptions{})
		if err != nil {
			t.Logf("failed to delete namespace %s: %v", tc.Namespace, err)
		}

		// Wait for namespace deletion
		deadline := time.Now().Add(30 * time.Second)
		for time.Now().Before(deadline) {
			_, err := tc.Client.CoreV1().Namespaces().Get(ctx, tc.Namespace, metav1.GetOptions{})
			if err != nil {
				break // Namespace deleted
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// CreateNamespace creates a namespace and registers it for cleanup
func (tc *TestCluster) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"test": "true",
			},
		},
	}

	created, err := tc.Client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	// Register cleanup
	tc.Cleanup = append(tc.Cleanup, func() error {
		return tc.Client.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	})

	return created, nil
}

// WaitForPodReady waits for pod to be ready
func (tc *TestCluster) WaitForPodReady(ctx context.Context, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pod, err := tc.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get pod: %w", err)
		}

		if pod.Status.Phase == corev1.PodRunning {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return nil
				}
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for pod %s to be ready", name)
}

// WaitForPVCBound waits for PVC to be bound
func (tc *TestCluster) WaitForPVCBound(ctx context.Context, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		pvc, err := tc.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get PVC: %w", err)
		}

		if pvc.Status.Phase == corev1.ClaimBound {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for PVC %s to be bound", name)
}

// sanitizeNamespaceName ensures namespace name follows K8s naming rules
func sanitizeNamespaceName(name string) string {
	// K8s namespace names must be DNS-1123 label
	// lowercase alphanumeric, max 63 chars
	if len(name) > 63 {
		name = name[:63]
	}
	// Replace invalid chars with dashes
	result := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			result += string(ch)
		} else {
			result += "-"
		}
	}
	return result
}
