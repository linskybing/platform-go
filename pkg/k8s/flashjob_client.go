package k8s

import (
	"context"
	"fmt"
	"time"

	k8stypes "github.com/linskybing/platform-go/pkg/k8s/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var flashJobGVR = schema.GroupVersionResource{
	Group:    "scheduling.flash-sched.io",
	Version:  "v1alpha1",
	Resource: "flashjobs",
}

// FlashJobClient wraps dynamic client operations for FlashJob CRDs.
type FlashJobClient struct {
	dynamicClient dynamic.Interface
}

func NewFlashJobClient(dynamicClient dynamic.Interface) *FlashJobClient {
	return &FlashJobClient{dynamicClient: dynamicClient}
}

func (c *FlashJobClient) Create(ctx context.Context, namespace string, job *k8stypes.FlashJob) (*k8stypes.FlashJob, error) {
	if c == nil || c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(job)
	if err != nil {
		return nil, fmt.Errorf("convert flashjob to unstructured: %w", err)
	}
	created, err := c.dynamicClient.Resource(flashJobGVR).Namespace(namespace).Create(ctx, &unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return unstructuredToFlashJob(created)
}

func (c *FlashJobClient) Get(ctx context.Context, namespace, name string) (*k8stypes.FlashJob, error) {
	if c == nil || c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	obj, err := c.dynamicClient.Resource(flashJobGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return unstructuredToFlashJob(obj)
}

func (c *FlashJobClient) Delete(ctx context.Context, namespace, name string) error {
	if c == nil || c.dynamicClient == nil {
		return fmt.Errorf("dynamic client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.dynamicClient.Resource(flashJobGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *FlashJobClient) Watch(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	if c == nil || c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if namespace == "" {
		return c.dynamicClient.Resource(flashJobGVR).Watch(ctx, opts)
	}
	return c.dynamicClient.Resource(flashJobGVR).Namespace(namespace).Watch(ctx, opts)
}

func (c *FlashJobClient) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*k8stypes.FlashJobList, error) {
	if c == nil || c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var obj *unstructured.UnstructuredList
	var err error
	if namespace == "" {
		obj, err = c.dynamicClient.Resource(flashJobGVR).List(ctx, opts)
	} else {
		obj, err = c.dynamicClient.Resource(flashJobGVR).Namespace(namespace).List(ctx, opts)
	}
	if err != nil {
		return nil, err
	}
	return unstructuredListToFlashJobList(obj)
}

func unstructuredToFlashJob(obj *unstructured.Unstructured) (*k8stypes.FlashJob, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil flashjob object")
	}
	var job k8stypes.FlashJob
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func unstructuredListToFlashJobList(obj *unstructured.UnstructuredList) (*k8stypes.FlashJobList, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil flashjob list")
	}
	var list k8stypes.FlashJobList
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &list); err != nil {
		return nil, err
	}
	return &list, nil
}
