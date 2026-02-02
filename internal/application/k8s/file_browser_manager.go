package k8s

import (
	"context"
	"fmt"

	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FileBrowserManager struct{}

func NewFileBrowserManager() *FileBrowserManager {
	return &FileBrowserManager{}
}

func (m *FileBrowserManager) GetPVCNames(ctx context.Context, namespace string) ([]string, error) {
	if k8sclient.Clientset == nil {
		return []string{}, nil
	}

	list, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "storage-type=project",
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(list.Items))
	for _, pvc := range list.Items {
		names = append(names, pvc.Name)
	}

	if len(names) == 0 {
		fallback, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
		if err == nil && len(fallback.Items) > 0 {
			names = make([]string, 0, len(fallback.Items))
			for _, pvc := range fallback.Items {
				names = append(names, pvc.Name)
			}
		}
	}

	return names, nil
}

func (m *FileBrowserManager) Start(ctx context.Context, ns string, pvcNames []string, readOnly bool, baseURL string) (string, error) {
	if len(pvcNames) == 0 {
		return "", fmt.Errorf("no PVCs available to start filebrowser")
	}

	_, err := k8sclient.CreateFileBrowserPod(ctx, ns, pvcNames, readOnly, baseURL)
	if err != nil {
		return "", err
	}

	nodePort, err := k8sclient.CreateFileBrowserService(ctx, ns)
	if err != nil {
		return "", err
	}

	return nodePort, nil
}

func (m *FileBrowserManager) Stop(ctx context.Context, ns string) error {
	return k8sclient.DeleteFileBrowserResources(ctx, ns)
}
