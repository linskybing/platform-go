package k8s

import (
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset     kubernetes.Interface
	clientsetOnce sync.Once
	useMock       bool
)

func init() {
	// Check if K8S_MOCK environment variable is set
	useMock = os.Getenv("K8S_MOCK") == "true"
}

// GetClientset returns the Kubernetes clientset (real or mock)
func GetClientset() kubernetes.Interface {
	clientsetOnce.Do(func() {
		if useMock {
			// Use fake clientset for testing
			clientset = fake.NewSimpleClientset()
		} else {
			// Use real Kubernetes client
			var err error
			clientset, err = initRealClient()
			if err != nil {
				// Fallback to mock if real client fails
				clientset = fake.NewSimpleClientset()
			}
		}
	})
	return clientset
}

// initRealClient initializes a real Kubernetes client
func initRealClient() (kubernetes.Interface, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err != nil {
		// If not in cluster, try kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(config)
}

// ResetClientForTesting resets the clientset (for testing only)
func ResetClientForTesting() {
	clientsetOnce = sync.Once{}
	clientset = nil
}

// SetMockMode sets whether to use mock client (for testing only)
func SetMockMode(mock bool) {
	useMock = mock
	ResetClientForTesting()
}
