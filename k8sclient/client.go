package k8sclient

import (
	"context"
	"encoding/json"
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
	v1 "k8s.io/api/core/v1"
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
	conn        *websocket.Conn
	stdinPipe   *io.PipeReader // 來自 websocket 的數據 -> stdin
	stdinWriter *io.PipeWriter
	sizeChan    chan remotecommand.TerminalSize
	once        sync.Once
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"` // 用於 stdin/stdout
	Cols int    `json:"cols,omitempty"` // 用於 resize
	Rows int    `json:"rows,omitempty"` // 用於 resize
}

var (
	Config        *rest.Config
	Clientset     *kubernetes.Clientset
	Dc            *discovery.DiscoveryClient
	Resources     []*restmapper.APIGroupResources
	Mapper        meta.RESTMapper
	DynamicClient *dynamic.DynamicClient
)

func InitTestCluster() {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		log.Fatal("KUBECONFIG is not set, cannot initialize test cluster")
	}

	var err error
	Config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("failed to load kubeconfig: %v", err)
	}

	Clientset, err = kubernetes.NewForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	server := Config.Host
	if !strings.Contains(server, "127.0.0.1") && !strings.Contains(server, "localhost") {
		log.Fatalf("unsafe cluster detected: %s, abort test", server)
	}

	Dc, err = discovery.NewDiscoveryClientForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create discovery client: %v", err)
	}

	Resources, err = restmapper.GetAPIGroupResources(Dc)
	if err != nil {
		log.Fatalf("failed to get API group resources: %v", err)
	}
	Mapper = restmapper.NewDiscoveryRESTMapper(Resources)

	DynamicClient, err = dynamic.NewForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create dynamic client: %v", err)
	}
}

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

// NewWebSocketIO 建立一個新的 WebSocketIO 處理器
func NewWebSocketIO(conn *websocket.Conn) *WebSocketIO {
	pr, pw := io.Pipe()
	handler := &WebSocketIO{
		conn:        conn,
		stdinPipe:   pr,
		stdinWriter: pw,
		sizeChan:    make(chan remotecommand.TerminalSize),
	}
	// 這裡不變，依然啟動 readLoop
	go handler.readLoop()
	return handler
}

// Read 方法會從接收 stdin 數據的管道中讀取數據 (實現 io.Reader)
func (h *WebSocketIO) Read(p []byte) (n int, err error) {
	return h.stdinPipe.Read(p)
}

// Write 方法會將數據寫入 WebSocket (實現 io.Writer)
func (h *WebSocketIO) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(TerminalMessage{
		Type: "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}
	if err := h.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Next 方法會被 executor 呼叫，用來等待一個 resize 事件 (實現 remotecommand.TerminalSizeQueue)
func (h *WebSocketIO) Next() *remotecommand.TerminalSize {
	size, ok := <-h.sizeChan
	if !ok {
		return nil // Channel 被關閉
	}
	return &size
}

// Close 用於清理資源
func (h *WebSocketIO) Close() {
	h.once.Do(func() {
		_ = h.stdinWriter.Close()
		close(h.sizeChan)
	})
}

// readLoop 是核心邏輯，在背景持續讀取 WebSocket 訊息並分發
func (h *WebSocketIO) readLoop() {
	// ✅ **核心修正：由 readLoop 自己負責清理**
	// 當這個 goroutine 退出時（無論是正常結束還是出錯），
	// defer 會確保 Close() 被呼叫，安全地關閉 channels。
	defer h.Close()

	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			// 當讀取發生錯誤 (例如 WebSocket 關閉)，
			// for 迴圈會終止，然後上面的 defer h.Close() 就會執行。
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "stdin":
			if msg.Data != "" {
				// 在這裡，因為迴圈還在繼續，所以 channel 肯定是打開的。
				_, _ = h.stdinWriter.Write([]byte(msg.Data))
			}
		case "resize":
			// 同上，channel 肯定是打開的。
			h.sizeChan <- remotecommand.TerminalSize{
				Width:  uint16(msg.Cols),
				Height: uint16(msg.Rows),
			}
		}
	}
}

// WebSocketIO's code remains the same, it's correct.
// ... NewWebSocketIO, Read, Write, Next, Close, readLoop ...

func ExecToPodViaWebSocket(
	conn *websocket.Conn,
	config *rest.Config,
	clientset *kubernetes.Clientset,
	namespace, podName, container string,
	command []string,
	tty bool,
) error {
	// Create our handler which implements all necessary interfaces.
	wsIO := NewWebSocketIO(conn)

	// ✅ **CORE FIX: Remove the defer from the main goroutine.**
	// The responsibility of closing the channels is now solely
	// within the readLoop goroutine. This eliminates the race condition.
	// defer wsIO.Close()  <-- REMOVE THIS LINE

	execCmd := []string{
		"env",
		"TERM=xterm",
	}
	execCmd = append(execCmd, command...) // Append the original command (e.g., "/bin/sh")

	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			// Use the modified command with the TERM variable.
			Command: execCmd,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     tty,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             wsIO,
		Stdout:            wsIO,
		Stderr:            wsIO,
		Tty:               tty,
		TerminalSizeQueue: wsIO,
	})
}

func WatchNamespaceResources(ctx context.Context, writeChan chan<- []byte, namespace string) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	var wg sync.WaitGroup
	for _, gvr := range gvrs {
		wg.Add(1)
		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchAndSend(ctx, DynamicClient, gvr, namespace, writeChan)
		}(gvr)
	}

	go func() {
		<-ctx.Done()
		wg.Wait()
		close(writeChan)
	}()
}

// WatchNamespaceResources 監控單一命名空間的資源變化
func WatchUserNamespaceResources(namespace string, writeChan chan<- []byte) {
	gvrs := []schema.GroupVersionResource{
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
	}

	// 同步等待所有資源監控結束
	var wg sync.WaitGroup

	// 為每個資源啟動一個監控協程
	for _, gvr := range gvrs {
		wg.Add(1)
		go func(gvr schema.GroupVersionResource) {
			defer wg.Done()
			watchUserAndSend(namespace, gvr, writeChan)
		}(gvr)
	}

	// 等待所有協程結束
	wg.Wait()
}

func watchUserAndSend(namespace string, gvr schema.GroupVersionResource, writeChan chan<- []byte) {
	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-time.After(time.Second * 10): // 增加超時避免死鎖
			return fmt.Errorf("timeout sending message")
		}
	}

	// initial list of resources
	list, err := DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			if err := sendObject("ADDED", &item); err != nil {
				fmt.Printf("Failed to send list item: %v\n", err)
			}
		}
	} else {
		fmt.Printf("List error for %s.%s: %v\n", gvr.Resource, gvr.Group, err)
	}

	// watch loop
	for {
		select {
		case <-time.After(time.Second * 30): // 進行每 30 秒重連
			// 每30秒進行一次 watch 重新連接
			watcher, err := DynamicClient.Resource(gvr).Namespace(namespace).Watch(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to start watch: %v\n", err)
				continue
			}

			for event := range watcher.ResultChan() {
				if obj, ok := event.Object.(*unstructured.Unstructured); ok {
					if err := sendObject(string(event.Type), obj); err != nil {
						fmt.Printf("Failed to send watch event: %v\n", err)
					}
				}
			}
		}
	}
}

func watchAndSend(
	ctx context.Context,
	dynClient dynamic.Interface,
	gvr schema.GroupVersionResource,
	ns string,
	writeChan chan<- []byte,
) {
	sendObject := func(eventType string, obj *unstructured.Unstructured) error {
		data := buildDataMap(eventType, obj)
		msg, err := json.Marshal(data)
		if err != nil {
			return err
		}

		select {
		case writeChan <- msg:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// initial list
	list, err := dynClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			if err := sendObject("ADDED", &item); err != nil {
				fmt.Printf("Failed to send list item: %v\n", err)
			}
		}
	} else {
		fmt.Printf("List error for %s.%s: %v\n", gvr.Resource, gvr.Group, err)
	}

	// watch loop
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

		for {
			select {
			case <-ctx.Done():
				watcher.Stop()
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}

				obj, ok := event.Object.(*unstructured.Unstructured)
				if !ok {
					continue
				}

				if err := sendObject(string(event.Type), obj); err != nil {
					fmt.Printf("Failed to send watch event: %v\n", err)
				}
			}
		}
	}
}
func buildDataMap(eventType string, obj *unstructured.Unstructured) map[string]interface{} {
	data := map[string]interface{}{
		"type": eventType,
		"kind": obj.GetKind(),
		"name": obj.GetName(),
		"ns":   obj.GetNamespace(),
	}

	for k, v := range extractStatusFields(obj) {
		data[k] = v
	}

	if obj.GetKind() == "Pod" {
		if containers, found, _ := unstructured.NestedSlice(obj.Object, "spec", "containers"); found {
			var containerNames []string
			for _, c := range containers {
				if m, ok := c.(map[string]interface{}); ok {
					if name, ok := m["name"].(string); ok {
						containerNames = append(containerNames, name)
					}
				}
			}
			if len(containerNames) > 0 {
				data["containers"] = containerNames
			}
		}
	}

	if isService(obj) {
		if ips := extractServiceExternalIPs(obj); len(ips) > 0 {
			data["externalIPs"] = ips
		}
		if ports := extractServiceNodePorts(obj); len(ports) > 0 {
			data["nodePorts"] = ports
		}
	}

	return data
}

func isService(obj *unstructured.Unstructured) bool {
	return obj.GetKind() == "Service"
}

func extractServiceExternalIPs(obj *unstructured.Unstructured) []string {
	var externalIPs []string

	specExternalIPs, found, err := unstructured.NestedSlice(obj.Object, "spec", "externalIPs")
	if found && err == nil {
		for _, ip := range specExternalIPs {
			if s, ok := ip.(string); ok {
				externalIPs = append(externalIPs, s)
			}
		}
	}

	ingressList, found, err := unstructured.NestedSlice(obj.Object, "status", "loadBalancer", "ingress")
	if found && err == nil {
		for _, ingress := range ingressList {
			if m, ok := ingress.(map[string]interface{}); ok {
				if ip, ok := m["ip"].(string); ok {
					externalIPs = append(externalIPs, ip)
				}
			}
		}
	}

	return externalIPs
}

func extractServiceNodePorts(obj *unstructured.Unstructured) []int64 {
	var nodePorts []int64

	ports, found, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
	if !found || err != nil {
		return nodePorts
	}

	for _, port := range ports {
		if m, ok := port.(map[string]interface{}); ok {
			if np, ok := m["nodePort"].(int64); ok {
				nodePorts = append(nodePorts, np)
			} else if npf, ok := m["nodePort"].(float64); ok {
				nodePorts = append(nodePorts, int64(npf))
			}
		}
	}

	return nodePorts
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

func GetFilteredNamespaces(filter string) ([]v1.Namespace, error) {
	namespaces, err := Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var filteredNamespaces []v1.Namespace
	for _, ns := range namespaces.Items {
		if strings.Contains(ns.Name, filter) {
			filteredNamespaces = append(filteredNamespaces, ns)
		}
	}

	return filteredNamespaces, nil
}
