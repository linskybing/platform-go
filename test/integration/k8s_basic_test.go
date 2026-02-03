package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/linskybing/platform-go/test/integration/k8stest"
)

func TestK8sNamespaceOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping K8s integration test in short mode")
	}

	tc := k8stest.SetupTestCluster(t)
	ctx := context.Background()

	t.Run("create namespace", func(t *testing.T) {
		nsName := "test-create-ns"
		ns, err := tc.CreateNamespace(ctx, nsName)
		require.NoError(t, err)
		assert.Equal(t, nsName, ns.Name)
		assert.Equal(t, "true", ns.Labels["test"])

		// Verify namespace exists
		retrieved, err := tc.Client.CoreV1().Namespaces().Get(ctx, nsName, metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, nsName, retrieved.Name)
	})

	t.Run("list namespaces with label selector", func(t *testing.T) {
		nsList, err := tc.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
			LabelSelector: "test=true",
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(nsList.Items), 1)
	})
}

func TestK8sPodOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping K8s integration test in short mode")
	}

	tc := k8stest.SetupTestCluster(t)
	ctx := context.Background()

	t.Run("create and wait for pod ready", func(t *testing.T) {
		pod := k8stest.CreateTestPod(tc.Namespace, "test-pod", "busybox:latest")

		created, err := tc.Client.CoreV1().Pods(tc.Namespace).Create(ctx, pod, metav1.CreateOptions{})
		require.NoError(t, err)
		assert.Equal(t, "test-pod", created.Name)

		// Wait for pod to be ready
		err = tc.WaitForPodReady(ctx, tc.Namespace, "test-pod", 60*time.Second)
		require.NoError(t, err)

		// Verify pod is running
		runningPod, err := tc.Client.CoreV1().Pods(tc.Namespace).Get(ctx, "test-pod", metav1.GetOptions{})
		require.NoError(t, err)
		assert.Equal(t, corev1.PodRunning, runningPod.Status.Phase)
	})

	t.Run("list pods with label selector", func(t *testing.T) {
		pods, err := tc.Client.CoreV1().Pods(tc.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "test=true",
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pods.Items), 1)
	})

	t.Run("delete pod", func(t *testing.T) {
		err := tc.Client.CoreV1().Pods(tc.Namespace).Delete(ctx, "test-pod", metav1.DeleteOptions{})
		require.NoError(t, err)

		// Verify pod is deleted
		_, err = tc.Client.CoreV1().Pods(tc.Namespace).Get(ctx, "test-pod", metav1.GetOptions{})
		assert.Error(t, err)
	})
}

func TestK8sPVCOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping K8s integration test in short mode")
	}

	tc := k8stest.SetupTestCluster(t)
	ctx := context.Background()

	t.Run("create PVC", func(t *testing.T) {
		pvc := k8stest.CreateTestPVC(tc.Namespace, "test-pvc", "standard", "1Gi")

		created, err := tc.Client.CoreV1().PersistentVolumeClaims(tc.Namespace).Create(ctx, pvc, metav1.CreateOptions{})
		require.NoError(t, err)
		assert.Equal(t, "test-pvc", created.Name)
		assert.Equal(t, "true", created.Labels["test"])
	})

	t.Run("list PVCs with label selector", func(t *testing.T) {
		pvcs, err := tc.Client.CoreV1().PersistentVolumeClaims(tc.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "test=true",
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(pvcs.Items), 1)
	})

	t.Run("update PVC labels", func(t *testing.T) {
		pvc, err := tc.Client.CoreV1().PersistentVolumeClaims(tc.Namespace).Get(ctx, "test-pvc", metav1.GetOptions{})
		require.NoError(t, err)

		pvc.Labels["updated"] = "true"
		updated, err := tc.Client.CoreV1().PersistentVolumeClaims(tc.Namespace).Update(ctx, pvc, metav1.UpdateOptions{})
		require.NoError(t, err)
		assert.Equal(t, "true", updated.Labels["updated"])
	})

	t.Run("delete PVC", func(t *testing.T) {
		err := tc.Client.CoreV1().PersistentVolumeClaims(tc.Namespace).Delete(ctx, "test-pvc", metav1.DeleteOptions{})
		require.NoError(t, err)
	})
}

func TestK8sServiceOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping K8s integration test in short mode")
	}

	tc := k8stest.SetupTestCluster(t)
	ctx := context.Background()

	t.Run("create service", func(t *testing.T) {
		svc := k8stest.CreateTestService(tc.Namespace, "test-svc", 8080)

		created, err := tc.Client.CoreV1().Services(tc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
		require.NoError(t, err)
		assert.Equal(t, "test-svc", created.Name)
		assert.Equal(t, corev1.ServiceTypeClusterIP, created.Spec.Type)
		assert.Equal(t, int32(8080), created.Spec.Ports[0].Port)
	})

	t.Run("get service", func(t *testing.T) {
		svc, err := tc.Client.CoreV1().Services(tc.Namespace).Get(ctx, "test-svc", metav1.GetOptions{})
		require.NoError(t, err)
		assert.NotEmpty(t, svc.Spec.ClusterIP)
	})

	t.Run("delete service", func(t *testing.T) {
		err := tc.Client.CoreV1().Services(tc.Namespace).Delete(ctx, "test-svc", metav1.DeleteOptions{})
		require.NoError(t, err)
	})
}
