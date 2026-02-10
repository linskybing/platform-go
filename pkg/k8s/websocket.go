package k8s

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// WebSocketIO handles the conversion between WebSocket and io.Reader/Writer
// It implements remotecommand.TerminalSizeQueue, io.Reader, and io.Writer
type WebSocketIO struct {
	conn        *websocket.Conn
	stdinPipe   *io.PipeReader
	stdinWriter *io.PipeWriter
	sizeChan    chan remotecommand.TerminalSize
	once        sync.Once
	mu          sync.Mutex // Protects concurrent writes (Ping vs Stdout)
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"` // For stdin/stdout
	Cols int    `json:"cols,omitempty"` // For resize
	Rows int    `json:"rows,omitempty"` // For resize
}

// NewWebSocketIO creates a new WebSocketIO handler and starts loops
func NewWebSocketIO(conn *websocket.Conn) *WebSocketIO {
	pr, pw := io.Pipe()

	// Context for internal coordination
	// ctx, cancel := context.WithCancel(context.Background())

	handler := &WebSocketIO{
		conn:        conn,
		stdinPipe:   pr,
		stdinWriter: pw,
		sizeChan:    make(chan remotecommand.TerminalSize),
		// cancel:      cancel,
	}

	// Start the main read loop (Standard Input from user)
	go handler.readLoop()
	// Start the ping loop (Heartbeat to client)
	go handler.pingLoop()

	return handler
}

// pingLoop sends periodic pings to keep the connection alive
func (h *WebSocketIO) pingLoop() {
	// Must be shorter than pongWait (60s)
	pingPeriod := 50 * time.Second
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		// Lock before writing to prevent race condition with stdout
		h.mu.Lock()
		// Set write deadline to prevent hanging
		if err := h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			h.mu.Unlock()
			h.Close()
			return
		}
		err := h.conn.WriteMessage(websocket.PingMessage, nil)
		h.mu.Unlock()

		if err != nil {
			// If ping fails, connection is likely dead. Close handler.
			h.Close()
			return
		}
	}
}

// Read reads data from the pipe receiving stdin data (implements io.Reader)
func (h *WebSocketIO) Read(p []byte) (n int, err error) {
	return h.stdinPipe.Read(p)
}

// Write writes data to WebSocket (stdout from Pod)
func (h *WebSocketIO) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(TerminalMessage{
		Type: "stdout",
		Data: string(p),
	})
	if err != nil {
		return 0, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Update WriteDeadline before writing
	_ = h.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := h.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Next is called by executor to wait for a resize event (implements remotecommand.TerminalSizeQueue)
func (h *WebSocketIO) Next() *remotecommand.TerminalSize {
	size, ok := <-h.sizeChan
	if !ok {
		return nil // Channel closed
	}
	return &size
}

// Close cleans up resources
// IMPORTANT: This method does NOT close sizeChan to avoid panics.
// sizeChan is closed by readLoop.
func (h *WebSocketIO) Close() {
	h.once.Do(func() {
		// Close stdinWriter to stop Read() calls if any
		_ = h.stdinWriter.Close()
		// We do NOT close sizeChan here because readLoop might be trying to send to it.
		// We do NOT close conn here immediately, we let readLoop handle the socket closure or wait for error.
	})
}

// readLoop is the core logic, continuously reading WebSocket messages in the background
func (h *WebSocketIO) readLoop() {
	// Cleanup when the loop exits
	defer func() {
		h.Close()          // Close pipes
		close(h.sizeChan)  // Close channel safely (ONLY here)
		_ = h.conn.Close() // Ensure underlying TCP connection is closed
	}()

	const pongWait = 60 * time.Second

	h.conn.SetReadLimit(512 * 1024)
	if err := h.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	h.conn.SetPongHandler(func(string) error {
		return h.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		// Refresh deadline on any message
		if err := h.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "stdin":
			if msg.Data != "" {
				if _, err := h.stdinWriter.Write([]byte(msg.Data)); err != nil {
					log.Printf("Failed to write to stdin: %v", err)
				}
			}
		case "resize":
			// Non-blocking send to avoid hanging if SPDY executor isn't ready
			// and to avoid panic if sizeChan is closed (though with defer structure it should be safe)
			select {
			case h.sizeChan <- remotecommand.TerminalSize{
				Width:  uint16(msg.Cols),
				Height: uint16(msg.Rows),
			}:
			default:
				// Drop resize event if channel buffer is full or no one listening
			}
		}
	}
}

func ExecToPodViaWebSocket(
	conn *websocket.Conn,
	config *rest.Config,
	clientset *kubernetes.Clientset,
	namespace, podName, container string,
	command []string,
	tty bool,
) error {
	wsIO := NewWebSocketIO(conn)

	// DO NOT call defer wsIO.Close() here.
	// Lifecycle is managed by NewWebSocketIO's goroutines.

	execCmd := []string{
		"env",
		"TERM=xterm",
	}
	execCmd = append(execCmd, command...)

	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   execCmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       tty,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	// This blocks until the command finishes
	return executor.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:             wsIO,
		Stdout:            wsIO,
		Stderr:            wsIO,
		Tty:               tty,
		TerminalSizeQueue: wsIO,
	})
}
