package k8sclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

type WebSocketIO struct {
	Conn       *websocket.Conn
	readBuffer chan []byte
	closeCh    chan struct{}

	sizeMu     sync.Mutex
	size       remotecommand.TerminalSize
	notifySize chan struct{}
}

var (
	Config        *rest.Config
	Clientset     *kubernetes.Clientset
	Dc            *discovery.DiscoveryClient
	Resources     []*restmapper.APIGroupResources
	Mapper        meta.RESTMapper
	DynamicClient *dynamic.DynamicClient
)

// Init 載入 kubeconfig，初始化 Clientset
func Init() {
	var err error
	if configPath := os.Getenv("KUBECONFIG"); configPath != "" {
		Config, err = clientcmd.BuildConfigFromFlags("", configPath)
	} else {
		Config, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
			Config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
	}
	if err != nil {
		log.Fatalf("failed to load kube config: %v", err)
	}
	Clientset, err = kubernetes.NewForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create kubernetes clientset: %v", err)
	}
	Dc, err = discovery.NewDiscoveryClientForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create Discovery client: %v", err)
	}
	Resources, err = restmapper.GetAPIGroupResources(Dc)
	if err != nil {
		log.Fatalf("failed to get api group resources: %v", err)
	}
	Mapper = restmapper.NewDiscoveryRESTMapper(Resources)
	DynamicClient, err = dynamic.NewForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create dynamic client: %v", err)
	}
}

func NewWebSocketIO(conn *websocket.Conn, initialCols, initialRows uint16) *WebSocketIO {
	wsio := &WebSocketIO{
		Conn:       conn,
		readBuffer: make(chan []byte, 10),
		closeCh:    make(chan struct{}),

		size:       remotecommand.TerminalSize{Width: initialCols, Height: initialRows},
		notifySize: make(chan struct{}, 1),
	}

	go func() {
		defer close(wsio.readBuffer)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var resize struct {
				Type string `json:"type"`
				Cols uint16 `json:"cols"`
				Rows uint16 `json:"rows"`
			}

			if err := json.Unmarshal(msg, &resize); err == nil && resize.Type == "resize" {
				wsio.sizeMu.Lock()
				wsio.size.Width = resize.Cols
				wsio.size.Height = resize.Rows
				wsio.sizeMu.Unlock()

				// 非阻塞通知有 resize
				select {
				case wsio.notifySize <- struct{}{}:
				default:
				}
				continue
			}

			// 不是 resize 就當一般輸入傳給 shell
			wsio.readBuffer <- msg
		}
	}()

	return wsio
}

// Read 實作 io.Reader，提供給 executor stdin
func (w *WebSocketIO) Read(p []byte) (int, error) {
	select {
	case b, ok := <-w.readBuffer:
		if !ok {
			return 0, io.EOF
		}
		return copy(p, b), nil
	case <-w.closeCh:
		return 0, errors.New("connection closed")
	}
}

// Write 實作 io.Writer，提供給 executor stdout/stderr
func (w *WebSocketIO) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Close 關閉連線
func (w *WebSocketIO) Close() error {
	close(w.closeCh)
	return w.Conn.Close()
}

// TerminalSizeQueue 介面實作，executor 呼叫這裡取得最新大小
func (w *WebSocketIO) Next() *remotecommand.TerminalSize {
	<-w.notifySize
	w.sizeMu.Lock()
	defer w.sizeMu.Unlock()
	return &w.size
}

// ExecToPodViaWebSocket 改用 WebSocketIO
func ExecToPodViaWebSocket(
	conn *websocket.Conn,
	config *rest.Config,
	clientset *kubernetes.Clientset,
	namespace, podName, container string,
	command []string,
	tty bool,
) error {
	wsIO := NewWebSocketIO(conn, 80, 24) // 初始 80x24

	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       tty,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             wsIO,
		Stdout:            wsIO,
		Stderr:            wsIO,
		Tty:               tty,
		TerminalSizeQueue: wsIO,
	})

	wsIO.Close()
	return err
}

func WatchNamespaceResources(conn *websocket.Conn, namespace string) {
	ctx := context.Background()

	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	for _, gvr := range gvrs {
		go watchAndSend(ctx, DynamicClient, gvr, namespace, conn)
	}
}

func watchAndSend(ctx context.Context, dynClient dynamic.Interface, gvr schema.GroupVersionResource, ns string, conn *websocket.Conn) {
	list, err := dynClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			data := map[string]interface{}{
				"type": "ADDED", // 模擬 informer 的 ADDED 事件
				"kind": item.GetKind(),
				"name": item.GetName(),
				"ns":   item.GetNamespace(),
			}
			for k, v := range extractStatusFields(&item) {
				data[k] = v
			}
			msg, _ := json.Marshal(data)
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	} else {
		// List 失敗也不阻止，繼續 Watch
		fmt.Printf("List error for %s.%s: %v\n", gvr.Resource, gvr.Group, err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		watcher, err := dynClient.Resource(gvr).Namespace(ns).Watch(ctx, metav1.ListOptions{})
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for event := range watcher.ResultChan() {
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			data := map[string]interface{}{
				"type": "ADDED", // 模擬 informer 的 ADDED 事件
				"kind": obj.GetKind(),
				"name": obj.GetName(),
				"ns":   obj.GetNamespace(),
			}
			for k, v := range extractStatusFields(obj) {
				data[k] = v
			}
			msg, _ := json.Marshal(data)
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func getWatchableNamespacedResources(dc *discovery.DiscoveryClient) ([]schema.GroupVersionResource, error) {
	apiResourceLists, err := dc.ServerPreferredNamespacedResources()
	if err != nil {
		return nil, err
	}

	var result []schema.GroupVersionResource
	for _, apiList := range apiResourceLists {
		gv, err := schema.ParseGroupVersion(apiList.GroupVersion)
		if err != nil {
			continue
		}
		for _, r := range apiList.APIResources {
			if r.Namespaced && contains(r.Verbs, "watch") && !strings.Contains(r.Name, "/") {
				result = append(result, schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: r.Name,
				})
			}
		}
	}
	return result, nil
}

func contains(sl []string, s string) bool {
	for _, item := range sl {
		if item == s {
			return true
		}
	}
	return false
}

func extractStatusFields(obj *unstructured.Unstructured) map[string]interface{} {
	kind := obj.GetKind()
	result := map[string]interface{}{}

	switch kind {
	case "Pod":
		if phase, found, _ := unstructured.NestedString(obj.Object, "status", "phase"); found {
			result["status"] = phase
		}
	case "Service":
		if clusterIP, found, _ := unstructured.NestedString(obj.Object, "spec", "clusterIP"); found {
			result["clusterIP"] = clusterIP
		}
		if externalIPs, found, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); found && len(externalIPs) > 0 {
			if ingressMap, ok := externalIPs[0].(map[string]interface{}); ok {
				if ip, ok := ingressMap["ip"].(string); ok {
					result["externalIP"] = ip
				}
				if hostname, ok := ingressMap["hostname"].(string); ok {
					result["externalHostname"] = hostname
				}
			}
		}
	case "Ingress":
		if externalIPs, found, _ := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress"); found && len(externalIPs) > 0 {
			if ingressMap, ok := externalIPs[0].(map[string]interface{}); ok {
				if ip, ok := ingressMap["ip"].(string); ok {
					result["externalIP"] = ip
				}
				if hostname, ok := ingressMap["hostname"].(string); ok {
					result["externalHostname"] = hostname
				}
			}
		}
	case "Deployment", "ReplicaSet":
		if availableReplicas, found, _ := unstructured.NestedInt64(obj.Object, "status", "availableReplicas"); found {
			result["availableReplicas"] = availableReplicas
		}
	case "Job":
		if succeeded, found, _ := unstructured.NestedInt64(obj.Object, "status", "succeeded"); found {
			result["succeeded"] = succeeded
		}
	}

	return result
}
