package image

import (
	"fmt"
	"strings"

	"github.com/linskybing/platform-go/internal/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ImageBuildTTLSeconds is the TTL (in seconds) for completed image pull jobs.
	ImageBuildTTLSeconds = 300
)

// BuildPullJob constructs a Kubernetes Job that pulls a source image and
// pushes it to the internal Harbor registry. It returns the Job object and
// the full source and harbor image references used in the job spec.
func BuildPullJob(name, tag string) (*batchv1.Job, string, string) {
	// Normalize name similar to previous implementation
	normalizedName := name
	parts := strings.Split(name, "/")

	hasDomain := strings.Contains(parts[0], ".") ||
		strings.Contains(parts[0], ":") ||
		parts[0] == "localhost"

	if !hasDomain {
		if len(parts) == 1 {
			normalizedName = "docker.io/library/" + name
		} else {
			normalizedName = "docker.io/" + name
		}
	}

	fullImage := fmt.Sprintf("%s:%s", normalizedName, tag)
	harborImage := fmt.Sprintf("%s%s:%s", config.HarborPrivatePrefix, name, tag)

	ttl := int32(ImageBuildTTLSeconds)

	k8sJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "image-puller-",
			Namespace:    "default",
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					ServiceAccountName: "default",
					InitContainers: []corev1.Container{
						{
							Name:            "pull-source",
							Image:           fullImage,
							ImagePullPolicy: corev1.PullAlways,
							Command:         []string{"/bin/sh", "-c", "echo 'Image pulled successfully'"},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "push-to-harbor",
							Image:           "gcr.io/go-containerregistry/crane:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         CraneCopyArgs(fullImage, harborImage),
							Env: []corev1.EnvVar{
								{
									Name:  "DOCKER_CONFIG",
									Value: "/kaniko/.docker",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "docker-config",
									MountPath: "/kaniko/.docker",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "docker-config",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "harbor-regcred",
									Items: []corev1.KeyToPath{
										{
											Key:  ".dockerconfigjson",
											Path: "config.json",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return k8sJob, fullImage, harborImage
}
