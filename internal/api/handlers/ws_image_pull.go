package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/pkg/response"
)

// WatchImagePullHandler monitors a specific image pull job status with production-grade error handling.
func WatchImagePullHandler(c *gin.Context, service *application.ImageService) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "job_id parameter is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger := slog.Default()
		logger.Error("websocket upgrade failed", "job_id", jobID, "error", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed"})
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Default().Debug("websocket close error", "job_id", jobID, "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	statusChan := service.SubscribeToPullJob(jobID)
	if statusChan == nil {
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"job not found"}`))
		return
	}

	conn.SetReadLimit(512)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	go func() {
		defer cancel()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

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
				slog.Default().Error("failed to marshal status", "job_id", jobID, "error", err)
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

// WatchMultiplePullJobsHandler monitors multiple pull job statuses with dynamic subscription.
// Optimized for production with resource cleanup and concurrent access control.
func WatchMultiplePullJobsHandler(c *gin.Context, service *application.ImageService) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger := slog.Default()
		logger.Error("websocket upgrade failed for multi-job watch", "error", err)
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "websocket upgrade failed"})
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Default().Debug("websocket close error for multi-job watch", "error", err)
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

	subscriptions := make(map[string]<-chan *application.PullJobStatus)
	jobIDsChan := make(chan string, 10)

	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	readDone := make(chan struct{})

	go func() {
		defer close(readDone)
		defer cancel()
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}

			if action, ok := msg["action"].(string); ok && action == "subscribe" {
				if jobID, ok := msg["job_id"].(string); ok && jobID != "" {
					select {
					case jobIDsChan <- jobID:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

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

		case <-readDone:
			return

		default:
			if len(subscriptions) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			for jobID, statusChan := range subscriptions {
				select {
				case status, ok := <-statusChan:
					if !ok {
						delete(subscriptions, jobID)
						continue
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
						slog.Default().Error("failed to marshal status for multi-job watch", "job_id", jobID, "error", err)
						continue
					}

					_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						return
					}
				default:
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}
