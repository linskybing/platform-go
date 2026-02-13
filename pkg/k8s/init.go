package k8s

import (
"log"
"os"
"strings"

"k8s.io/apimachinery/pkg/api/meta"
"k8s.io/client-go/discovery"
"k8s.io/client-go/dynamic"
"k8s.io/client-go/kubernetes"
"k8s.io/client-go/rest"
"k8s.io/client-go/restmapper"
)

// Global variables for Kubernetes client initialization
// Deprecated: Use Client struct instead.
var (
Config        *rest.Config
Clientset     kubernetes.Interface
Dc            *discovery.DiscoveryClient
Resources     []*restmapper.APIGroupResources // Kept for types, but might remain nil if not exposed by Manager
Mapper        meta.RESTMapper
DynamicClient *dynamic.DynamicClient
)

// Init initializes the Kubernetes client and populates global variables.
// Deprecated: Use NewClient() instead.
func Init() {
client, err := NewClient()
if err != nil {
log.Fatalf("failed to initialize kubernetes client: %v", err)
}

// Populate globals for backward compatibility
Config = client.Config
Clientset = client.Clientset
Dc = client.Discovery
Mapper = client.Mapper
DynamicClient = client.DynamicClient
}

// InitTestCluster initializes a fake or real client for testing.
func InitTestCluster() {
kubeconfig := os.Getenv("KUBECONFIG")
if kubeconfig == "" {
log.Println("KUBECONFIG is not set, using fake Kubernetes client for tests")
client := NewFakeClient()
Clientset = client.Clientset
Config = client.Config
return
}

// Reuse NewClient logic but with specific check
client, err := NewClient()
if err != nil {
log.Fatalf("failed to create clientset: %v", err)
}

server := client.Config.Host
if !strings.Contains(server, "127.0.0.1") && !strings.Contains(server, "localhost") {
log.Fatalf("unsafe cluster detected: %s, abort test", server)
}

// Populate globals
Config = client.Config
Clientset = client.Clientset
Dc = client.Discovery
Mapper = client.Mapper
DynamicClient = client.DynamicClient

// Re-fetch resources to populate the global Resources variable if absolutely needed
// purely for strict backward compatibility if any internal test relies on it.
Resources, _ = restmapper.GetAPIGroupResources(Dc)
}
