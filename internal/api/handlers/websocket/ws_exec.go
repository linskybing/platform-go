package websocket

import (
"encoding/json"
"net/http"

"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
"github.com/linskybing/platform-go/pkg/k8s"
"github.com/linskybing/platform-go/pkg/response"
"k8s.io/client-go/kubernetes"
)

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
