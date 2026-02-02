package k8s

import (
	"context"
	"fmt"

	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
)

type FileBrowserManager struct{}

func NewFileBrowserManager() *FileBrowserManager {
	return &FileBrowserManager{}
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
