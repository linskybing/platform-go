package k8s

import (
	"context"

	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (sm *StorageManager) ensureGroupNamespace(ctx context.Context, nsName string, groupID string) error {
	if k8sclient.Clientset == nil {
		return nil
	}

	_, err := k8sclient.Clientset.CoreV1().Namespaces().Get(ctx, nsName, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	newNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
			Labels: map[string]string{
				"managed-by": "platform",
				"type":       "group-storage",
				"group-id":   groupID,
			},
		},
	}

	_, err = k8sclient.Clientset.CoreV1().Namespaces().Create(ctx, newNs, metav1.CreateOptions{})
	return err
}
