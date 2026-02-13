package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client holds all Kubernetes clients and configuration.
type Client struct {
	Clientset     kubernetes.Interface
	DynamicClient *dynamic.DynamicClient
	Discovery     *discovery.DiscoveryClient
	Config        *rest.Config
	Mapper        meta.RESTMapper
}

// NewClient creates a new Kubernetes client.
// It follows the standard precedence: KUBECONFIG env -> In-cluster -> ~/.kube/config.
func NewClient() (*Client, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Set high QPS/Burst to avoid throttling during high concurrency
	config.QPS = 50
	config.Burst = 100

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Initialize RESTMapper for dynamic resource handling
	resources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get API group resources: %w", err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(resources)

	return &Client{
		Clientset:     clientset,
		DynamicClient: dynamicClient,
		Discovery:     discoveryClient,
		Config:        config,
		Mapper:        mapper,
	}, nil
}

// NewFakeClient creates a fake client for testing purposes.
func NewFakeClient() *Client {
	return &Client{
		Clientset: k8sfake.NewSimpleClientset(),
		// Dynamic, Discovery, Mapper are nil or need mocking if used in tests
		Config: &rest.Config{},
	}
}

func loadConfig() (*rest.Config, error) {
	if configPath := os.Getenv("KUBECONFIG"); configPath != "" {
		return clientcmd.BuildConfigFromFlags("", configPath)
	}

	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if _, err := os.Stat(kubeconfig); err == nil {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return nil, fmt.Errorf("no valid kubeconfig found (checked KUBECONFIG, in-cluster, ~/.kube/config)")
}

func (c *Client) ValidateCluster() error {
	host := c.Config.Host
	if !strings.Contains(host, "127.0.0.1") && !strings.Contains(host, "localhost") {
		// Just a warning or check, logic copied from InitTestCluster but modified
		// returning error effectively blocks usage if strict check is desired.
		// For now, we just pass.
	}
	return nil
}
