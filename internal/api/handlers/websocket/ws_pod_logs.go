package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// jobPodDiscoveryTimeout is the maximum time to wait for a job's pod to appear.
	jobPodDiscoveryTimeout = 60 * time.Second
	// jobPodDiscoveryInterval is the polling interval while waiting for a job pod.
	jobPodDiscoveryInterval = 2 * time.Second
)

// StreamPodLogsHandler streams pod logs over WebSocket with production-grade optimization.
// Query params: namespace (required), pod (optional), job (optional), container (optional), tailLines (optional).
// Implements timeout management and graceful error handling.
func StreamPodLogsHandler(c *gin.Context) {
	logger := slog.Default()

	namespace := c.Query("namespace")
	podName := c.Query("pod")
	jobName := c.Query("job")
	container := c.Query("container")

	if namespace == "" {
		response.Error(c, http.StatusBadRequest, "Namespace is required")
		return
	}
	if podName == "" && jobName == "" {
		response.Error(c, http.StatusBadRequest, "Pod or job is required")
		return
	}

	if k8s.Clientset == nil {
		response.Error(c, http.StatusServiceUnavailable, "K8s client not available")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("websocket upgrade failed", "namespace", namespace, "pod", podName, "job", jobName, "error", err)
		response.Error(c, http.StatusInternalServerError, "Websocket upgrade failed")
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Debug("websocket close error", "namespace", namespace, "pod", podName, "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Resolve pod name from job if needed
	if podName == "" {
		var err error
		podName, err = waitForJobPod(ctx, namespace, jobName, logger)
		if err != nil {
			logger.Error("failed to resolve pod from job", "namespace", namespace, "job", jobName, "error", err)
			writeLogError(conn, fmt.Sprintf("failed to resolve pod: %v", err))
			return
		}
	}

	// Parse tailLines parameter with production validation
	var tailLines *int64
	if raw := c.Query("tailLines"); raw != "" {
		if n := parsePositiveInt(raw); n > 0 && n <= 10000 {
			v := int64(n)
			tailLines = &v
		}
	}

	logReq := k8s.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: tailLines,
	})

	stream, err := logReq.Stream(ctx)
	if err != nil {
		logger.Error("failed to stream logs", "namespace", namespace, "pod", podName, "container", container, "error", err)
		writeLogError(conn, fmt.Sprintf("failed to stream logs: %v", err))
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			logger.Debug("stream close error", "namespace", namespace, "pod", podName, "error", err)
		}
	}()

	var writeMu sync.Mutex
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			writeMu.Lock()
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				writeMu.Unlock()
				cancel()
				return
			}
			writeMu.Unlock()
		}
	}()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		case <-readDone:
			return
		default:
			n, readErr := stream.Read(buf)
			if n > 0 {
				payload, err := json.Marshal(gin.H{
					"type": "log",
					"data": string(buf[:n]),
				})
				if err != nil {
					logger.Error("failed to marshal log message", "namespace", namespace, "pod", podName, "error", err)
					continue
				}

				writeMu.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
					writeMu.Unlock()
					logger.Error("failed to write log message", "namespace", namespace, "pod", podName, "error", err)
					return
				}
				writeMu.Unlock()
			}
			if readErr != nil {
				return
			}
		}
	}
}

func waitForJobPod(ctx context.Context, namespace, jobName string, logger *slog.Logger) (string, error) {
	if jobName == "" {
		return "", fmt.Errorf("job name is required")
	}

	ticker := time.NewTicker(jobPodDiscoveryInterval)
	defer ticker.Stop()

	timeout := time.NewTimer(jobPodDiscoveryTimeout)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context canceled while waiting for job pod")
		case <-timeout.C:
			logger.Warn("timeout waiting for job pod", "namespace", namespace, "job", jobName)
			return "", fmt.Errorf("timeout waiting for job pod after %v", jobPodDiscoveryTimeout)
		case <-ticker.C:
			pods, err := k8s.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
				LabelSelector: "job-name=" + jobName,
			})
			if err != nil {
				logger.Debug("failed to list pods for job", "namespace", namespace, "job", jobName, "error", err)
				continue
			}
			if len(pods.Items) > 0 {
				return pods.Items[0].Name, nil
			}
		}
	}
}

func writeLogError(conn *websocket.Conn, message string) {
	payload, _ := json.Marshal(gin.H{
		"type":  "error",
		"error": message,
	})
	_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = conn.WriteMessage(websocket.TextMessage, payload)
}

func parsePositiveInt(raw string) int {
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return 0
	}
	return val
}
