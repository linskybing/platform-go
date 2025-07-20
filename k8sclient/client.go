package k8sclient

import (
	"log"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

var (
	Config    *rest.Config
	Clientset *kubernetes.Clientset
)

func Init() {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	var err error
	Config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("failed to load kubeconfig: %v", err)
	}

	Clientset, err = kubernetes.NewForConfig(Config)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}
}

// WebSocketIO implements io.Reader and io.Writer for WebSocket
type WebSocketIO struct {
	Conn *websocket.Conn
}

func (w *WebSocketIO) Read(p []byte) (int, error) {
	_, msg, err := w.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, msg), nil
}

func (w *WebSocketIO) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Start WebSocket <-> Pod Exec streaming
func ExecToPodViaWebSocket(
	conn *websocket.Conn,
	config *rest.Config,
	clientset *kubernetes.Clientset,
	namespace, podName, container string,
	command []string,
	tty bool,
) error {
	wsIO := &WebSocketIO{Conn: conn}

	// Optional: ping every 30s to keep WebSocket alive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
			<-ticker.C
		}
	}()

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

	return executor.Stream(remotecommand.StreamOptions{
		Stdin:  wsIO,
		Stdout: wsIO,
		Stderr: wsIO,
		Tty:    tty,
	})
}
