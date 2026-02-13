package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// PodDeleteTimeout is the maximum time to wait for pod deletion.
	PodDeleteTimeout = 5 * time.Second
	// PodDeletePollInterval is the polling interval during pod deletion.
	PodDeletePollInterval = 200 * time.Millisecond
)

func GetFilteredNamespaces(filter string) ([]corev1.Namespace, error) {
	ctx, cancel := requestContext()
	defer cancel()
	namespaces, err := Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var filteredNamespaces []corev1.Namespace
	for _, ns := range namespaces.Items {
		if strings.Contains(ns.Name, filter) {
			filteredNamespaces = append(filteredNamespaces, ns)
		}
	}

	return filteredNamespaces, nil
}

type JobSpec struct {
	Name              string
	Namespace         string
	Image             string
	Command           []string
	PriorityClassName string
	Parallelism       int32
	Completions       int32
	Volumes           []VolumeSpec
	GPUCount          int
	GPUType           string
	CPURequest        string
	MemoryRequest     string
	EnvVars           map[string]string
	Annotations       map[string]string
}

type VolumeSpec struct {
	Name      string
	PVCName   string
	HostPath  string
	MountPath string
}

// CreateJob creates a Kubernetes Job with flexible configuration
func CreateJob(ctx context.Context, spec JobSpec) error {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount

	for _, v := range spec.Volumes {
		var volumeSource corev1.VolumeSource
		if v.PVCName != "" {
			volumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.PVCName,
				},
			}
		} else if v.HostPath != "" {
			volumeSource = corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: v.HostPath,
				},
			}
		}

		volumes = append(volumes, corev1.Volume{
			Name:         v.Name,
			VolumeSource: volumeSource,
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      v.Name,
			MountPath: v.MountPath,
		})
	}

	var env []corev1.EnvVar
	for k, v := range spec.EnvVars {
		env = append(env, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	container := corev1.Container{
		Name:         spec.Name,
		Image:        spec.Image,
		Command:      spec.Command,
		VolumeMounts: volumeMounts,
		Env:          env,
	}

	resources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if spec.GPUCount > 0 {
		qty := resource.MustParse(fmt.Sprintf("%d", spec.GPUCount))
		resourceName := corev1.ResourceName("nvidia.com/gpu")
		if spec.GPUType == "shared" {
			resourceName = corev1.ResourceName("nvidia.com/gpu.shared")
		}

		resources.Limits[resourceName] = qty
		resources.Requests[resourceName] = qty
	}

	if spec.CPURequest != "" {
		if q, err := resource.ParseQuantity(spec.CPURequest); err == nil {
			resources.Requests[corev1.ResourceCPU] = q
		}
	}

	if spec.MemoryRequest != "" {
		if q, err := resource.ParseQuantity(spec.MemoryRequest); err == nil {
			resources.Requests[corev1.ResourceMemory] = q
		}
	}

	container.Resources = resources

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: spec.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism: &spec.Parallelism,
			Completions: &spec.Completions,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: spec.Annotations,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:     corev1.RestartPolicyOnFailure,
					PriorityClassName: spec.PriorityClassName,
					Volumes:           volumes,
					Containers: []corev1.Container{
						container,
					},
				},
			},
		},
	}

	_, err := Clientset.BatchV1().Jobs(spec.Namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

// DeleteJob deletes a Kubernetes Job and its pods.
func DeleteJob(ctx context.Context, namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return Clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

// CreateFileBrowserPod creates a pod running filebrowser with multiple PVC mounts
func CreateFileBrowserPod(ctx context.Context, ns string, pvcNames []string, readOnly bool, baseURL string) (string, error) {
	if len(pvcNames) == 0 {
		return "", fmt.Errorf("no PVCs provided for filebrowser")
	}

	podName := "filebrowser-project"

	existingPod, err := Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err == nil {
		matches := true
		if len(existingPod.Spec.Containers) > 0 {
			for _, m := range existingPod.Spec.Containers[0].VolumeMounts {
				if m.ReadOnly != readOnly {
					matches = false
					break
				}
			}
		}

		if matches {
			return podName, nil
		}

		grace := int64(0)
		_ = Clientset.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{GracePeriodSeconds: &grace})
		_ = wait.PollUntilContextTimeout(ctx, PodDeletePollInterval, PodDeleteTimeout, true, func(ctx context.Context) (bool, error) {
			_, err := Clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		})
	}

	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount
	for idx, pvc := range pvcNames {
		volName := fmt.Sprintf("data-%d", idx)
		volumes = append(volumes, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: pvc},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: "/srv",
			ReadOnly:  readOnly,
		})
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
			Labels: map[string]string{
				"app":  "filebrowser",
				"role": "project-storage",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "filebrowser",
					Image: "filebrowser/filebrowser:latest",
					Args: []string{
						"--noauth",
						"--database", "/tmp/filebrowser.db",
						"--root", "/srv",
						"--port", "80",
						"--address", "0.0.0.0",
						"--baseURL", baseURL,
					},
					Ports:        []corev1.ContainerPort{{ContainerPort: 80}},
					VolumeMounts: mounts,
				},
			},
			Volumes: volumes,
		},
	}

	_, err = Clientset.CoreV1().Pods(ns).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return podName, nil
}

func CreateFileBrowserService(ctx context.Context, ns string) (string, error) {
	svcName := config.GroupStorageBrowserSVCName

	svc, err := Clientset.CoreV1().Services(ns).Get(ctx, svcName, metav1.GetOptions{})
	if err == nil {
		if len(svc.Spec.Ports) > 0 {
			return fmt.Sprintf("%d", svc.Spec.Ports[0].NodePort), nil
		}
		return "", nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "filebrowser",
				"role": "project-storage",
			},
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	createdSvc, err := Clientset.CoreV1().Services(ns).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	if len(createdSvc.Spec.Ports) > 0 {
		return fmt.Sprintf("%d", createdSvc.Spec.Ports[0].NodePort), nil
	}
	return "", nil
}

func DeleteFileBrowserResources(ctx context.Context, ns string) error {
	podName := "filebrowser-group"
	svcName := config.GroupStorageBrowserSVCName

	err := Clientset.CoreV1().Services(ns).Delete(ctx, svcName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	err = Clientset.CoreV1().Pods(ns).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}
