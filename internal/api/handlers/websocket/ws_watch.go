package websocket

import (
"context"
"log"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
"github.com/linskybing/platform-go/pkg/k8s"
"github.com/linskybing/platform-go/pkg/response"
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

// Reader Loop: Consume control messages (Ping/Pong/Close)
// Even if we don't expect client messages, we must read to process Ping/Pong/Close frames.
go func() {
defer cancel() // Cancel context if connection drops
for {
if _, _, err := conn.NextReader(); err != nil {
if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
log.Printf("error: %v", err)
}
break
}
}
}()

// Data Channel: K8s watcher writes raw JSON bytes here
// Buffer allows watcher to proceed even if writer is slightly behind
writeChan := make(chan []byte, 100)

// Start K8s Watcher (Background Producer)
go k8s.WatchUserNamespaceResources(ctx, namespace, writeChan)

// Writer Loop: Batches messages and sends to WebSocket (Consumer)
ticker := time.NewTicker(pingPeriod)
defer ticker.Stop()

// Flush ticker for batching
flushTicker := time.NewTicker(flushFrequency)
defer flushTicker.Stop()

var batchData [][]byte

for {
select {
case <-ctx.Done():
return

case <-flushTicker.C:
// Send buffered messages if any
if len(batchData) > 0 {
_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
// Wrap batch in array [msg1, msg2, ...]
// Since msg1 is already JSON bytes, we need to manually construct the array or use multiple writes.
// However, standard specificiation usually implies streaming events.
// If frontend expects array:
// combined := "[" + bytes.Join(batchData, []byte(",")) + "]"
// But here we rely on the implementation of WatchUserNamespaceResources which sends complete objects.
// Sending simple objects one by one is safer for compatibility unless frontend specifically supports batching.

// NOTE: Previous implementation might have assumed single messages.
// If we want batching, we should check frontend contract. 
// For now, let's just write them one by one to be safe, but quickly.

for _, msg := range batchData {
_ = conn.WriteMessage(websocket.TextMessage, msg)
}
batchData = nil // Reset buffer
}

case msg, ok := <-writeChan:
if !ok {
_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
return
}

batchData = append(batchData, msg)
if len(batchData) >= batchSize {
_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
for _, m := range batchData {
_ = conn.WriteMessage(websocket.TextMessage, m)
}
batchData = nil
}

case <-ticker.C:
_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
return
}
}
}
}
