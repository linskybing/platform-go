package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/response"
	"k8s.io/client-go/kubernetes"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	// Reduced to 60s to prevent Load Balancers (Nginx/AWS) from dropping idle connections.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Batching: Maximum number of messages to buffer before forcing a send
	batchSize = 50

	// Batching: Maximum time to wait before sending buffered messages
	flushFrequency = 100 * time.Millisecond
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now, consider restricting this in production
		return true
	},
}

// ExecWebSocketHandler handles "kubectl exec" style terminal sessions
func ExecWebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed: " + err.Error()})
		return
	}

	cs, ok := k8s.Clientset.(*kubernetes.Clientset)
	if !ok || cs == nil {
		c.JSON(http.StatusServiceUnavailable, response.ErrorResponse{Error: "k8s client not available"})
		return
	}

	// k8s.ExecToPodViaWebSocket typically manages its own stream copy loops
	err = k8s.ExecToPodViaWebSocket(
		conn,
		k8s.Config,
		cs,
		c.Query("namespace"),
		c.Query("pod"),
		c.Query("container"),
		[]string{c.DefaultQuery("command", "/bin/bash")},
		c.DefaultQuery("tty", "true") == "true",
	)

	if err != nil {
		errorMsg := k8s.TerminalMessage{
			Type: "stdout",
			Data: "\r\n\x1b[31m[Error] " + err.Error() + "\x1b[0m\r\n",
		}
		jsonMsg, _ := json.Marshal(errorMsg)
		_ = conn.WriteMessage(websocket.TextMessage, jsonMsg)
		_ = conn.Close()
		return
	}
}

// WatchNamespaceHandler monitors resources for a specific namespace
// Features: Heartbeat, Message Batching, Context Cancellation
func WatchNamespaceHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "namespace parameter is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed: " + err.Error()})
		return
	}

	// Context used to coordinate shutdown between Reader, Writer, and K8s Watcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure Heartbeat (Reader Side)
	conn.SetReadLimit(512 * 1024)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	writeChan := make(chan []byte, 200)

	// Writer Goroutine: Handles Batching, Pings, and sending data to client
	go func() {
		defer func() { _ = conn.Close() }()

		pingTicker := time.NewTicker(pingPeriod)
		defer pingTicker.Stop()

		flushTicker := time.NewTicker(flushFrequency)
		defer flushTicker.Stop()

		var buffer []json.RawMessage

		// Helper to flush buffer to WebSocket
		flush := func() error {
			if len(buffer) == 0 {
				return nil
			}
			batchData, err := json.Marshal(buffer)
			if err != nil {
				return err
			}
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, batchData); err != nil {
				return err
			}
			buffer = buffer[:0]
			return nil
		}

		for {
			select {
			case msg, ok := <-writeChan:
				if !ok {
					_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				buffer = append(buffer, json.RawMessage(msg))

				// Flush immediately if batch is full
				if len(buffer) >= batchSize {
					if err := flush(); err != nil {
						cancel()
						return
					}
				}

			case <-flushTicker.C:
				if err := flush(); err != nil {
					cancel()
					return
				}

			case <-pingTicker.C:
				// Flush pending data before sending Ping
				if err := flush(); err != nil {
					cancel()
					return
				}
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					cancel()
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	// Start K8s Watcher
	go k8s.WatchNamespaceResources(ctx, writeChan, namespace)

	// Reader Loop (Blocking)
	// Essential for processing Control Frames (Ping/Pong/Close)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// WatchImagePullHandler monitors a specific image pull job status
func WatchImagePullHandler(c *gin.Context, imageService interface{}) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "job_id parameter is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed: " + err.Error()})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service, ok := imageService.(*application.ImageService)
	if !ok {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"service not available"}`))
		return
	}

	statusChan := service.SubscribeToPullJob(jobID)
	if statusChan == nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"job not found"}`))
		return
	}

	// IMPORTANT: Heartbeat Configuration
	// We must set ReadDeadline and handle Pongs, otherwise the connection will rot.
	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Reader Goroutine
	// Required to consume control frames (Pong) sent by the client.
	// If we don't read, the TCP buffer fills up and the connection dies.
	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Writer Loop
	for {
		select {
		case status, ok := <-statusChan:
			if !ok {
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(gin.H{
				"job_id":    status.JobID,
				"image":     status.ImageName + ":" + status.ImageTag,
				"status":    status.Status,
				"progress":  status.Progress,
				"message":   status.Message,
				"timestamp": status.UpdatedAt,
			})
			if err != nil {
				continue
			}

			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// WatchMultiplePullJobsHandler monitors multiple pull job statuses with dynamic subscription
func WatchMultiplePullJobsHandler(c *gin.Context, imageService interface{}) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed: " + err.Error()})
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service, ok := imageService.(*application.ImageService)
	if !ok {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"service not available"}`))
		return
	}

	// Configure Heartbeat
	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	subscriptions := make(map[string]<-chan *application.PullJobStatus)
	jobIDsChan := make(chan string, 10)

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Reader Goroutine: Handles incoming JSON commands and Pongs
	go func() {
		defer cancel()
		for {
			var msg map[string]interface{}
			// ReadJSON internally calls ReadMessage, keeping the heartbeat alive
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}

			if action, ok := msg["action"].(string); ok && action == "subscribe" {
				if jobID, ok := msg["job_id"].(string); ok {
					select {
					case jobIDsChan <- jobID:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	// Writer Loop
	for {
		select {
		case jobID := <-jobIDsChan:
			if _, exists := subscriptions[jobID]; !exists {
				if statusChan := service.SubscribeToPullJob(jobID); statusChan != nil {
					subscriptions[jobID] = statusChan
				}
			}

		case <-pingTicker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-ctx.Done():
			return

		// Non-blocking check for subscription updates
		default:
			allClosed := true
			hasSubscriptions := len(subscriptions) > 0

			for jobID, statusChan := range subscriptions {
				select {
				case status, ok := <-statusChan:
					if !ok {
						delete(subscriptions, jobID)
						continue
					}
					allClosed = false

					data, err := json.Marshal(gin.H{
						"job_id":    status.JobID,
						"image":     status.ImageName + ":" + status.ImageTag,
						"status":    status.Status,
						"progress":  status.Progress,
						"message":   status.Message,
						"timestamp": status.UpdatedAt,
					})
					if err == nil {
						_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
						if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
							return
						}
					}
				default:
					// Channel has no data yet, continue to next subscription
					allClosed = false
				}
			}

			// Small sleep to prevent tight loop CPU spike if no data is available
			if hasSubscriptions && !allClosed {
				time.Sleep(100 * time.Millisecond)
			} else if !hasSubscriptions {
				// If no subscriptions, just wait for ticker or new jobs
				select {
				case jobID := <-jobIDsChan:
					if statusChan := service.SubscribeToPullJob(jobID); statusChan != nil {
						subscriptions[jobID] = statusChan
					}
				case <-pingTicker.C:
					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
