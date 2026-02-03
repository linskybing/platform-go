package k8stest

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTestPVC creates a PVC for testing
func CreateTestPVC(namespace, name, storageClass string, size string) *corev1.PersistentVolumeClaim {
	storageClassName := storageClass
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "true",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
}

// CreateTestPod creates a pod for testing
func CreateTestPod(namespace, name, image string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "true",
				"app":  name,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "test",
					Image:   image,
					Command: []string{"sleep", "3600"},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
	}
}

// CreateTestService creates a service for testing
func CreateTestService(namespace, name string, port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"test": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:     port,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// CleanupNamespaceResources deletes all test resources in namespace
func CleanupNamespaceResources(ctx context.Context, tc *TestCluster, namespace string) error {
	// Delete all pods
	err := tc.Client.CoreV1().Pods(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: "test=true"},
	)
	if err != nil {
		return err
	}

	// Delete all PVCs
	err = tc.Client.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{LabelSelector: "test=true"},
	)
	if err != nil {
		return err
	}

	// Delete all services individually (no DeleteCollection for services)
	services, err := tc.Client.CoreV1().Services(namespace).List(
		ctx,
		metav1.ListOptions{LabelSelector: "test=true"},
	)
	if err != nil {
		return err
	}

	for _, svc := range services.Items {
		err = tc.Client.CoreV1().Services(namespace).Delete(
			ctx,
			svc.Name,
			metav1.DeleteOptions{},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
