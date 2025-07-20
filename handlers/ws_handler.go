package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/linskybing/platform-go/k8sclient"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ExecWebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed: " + err.Error()})
		return
	}
	defer conn.Close()

	namespace := c.Query("namespace")
	pod := c.Query("pod")
	container := c.Query("container")
	tty := c.DefaultQuery("tty", "true") == "true"
	shell := c.DefaultQuery("command", "/bin/sh")

	command := []string{shell}
	err = k8sclient.ExecToPodViaWebSocket(
		conn,
		k8sclient.Config,
		k8sclient.Clientset,
		namespace,
		pod,
		container,
		command,
		tty,
	)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("exec error: "+err.Error()))
		conn.Close()
		return
	}
}
