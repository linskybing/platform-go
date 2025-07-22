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
	// 升級成功後，不要再用 c.Request.Body 或其他 HTTP 輸入
	// 交給 ExecToPodViaWebSocket 處理 websocket 連線

	err = k8sclient.ExecToPodViaWebSocket(
		conn,
		k8sclient.Config,
		k8sclient.Clientset,
		c.Query("namespace"),
		c.Query("pod"),
		c.Query("container"),
		[]string{c.DefaultQuery("command", "/bin/sh")},
		c.DefaultQuery("tty", "true") == "true",
	)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("exec error: "+err.Error()))
		conn.Close()
		return
	}
}

func WatchNamespaceHandler(c *gin.Context) {
	namespace := c.Param("namespace")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "websocket upgrade failed: " + err.Error()})
		return
	}

	go k8sclient.WatchNamespaceResources(conn, namespace)

	// 監聽 client 是否關閉連線
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			conn.Close()
			break
		}
	}
}
