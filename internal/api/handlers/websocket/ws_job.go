package websocket

import (
"encoding/json"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gorilla/websocket"
"github.com/linskybing/platform-go/internal/repository"
"github.com/linskybing/platform-go/pkg/response"
"log/slog"
)

// WatchJobStatusHandler streams status updates for a specific job
func WatchJobStatusHandler(c *gin.Context, repos *repository.Repos) {
jobID := c.Param("id")
if jobID == "" {
c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "job id is required"})
return
}

conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
if err != nil {
slog.Error("websocket upgrade failed", "job_id", jobID, "error", err)
return
}
defer conn.Close()

ctx := c.Request.Context()

// Initial status
j, err := repos.Job.Get(ctx, jobID)
if err != nil {
_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "job not found"}`))
return
}

initData, _ := json.Marshal(j)
_ = conn.WriteMessage(websocket.TextMessage, initData)

ticker := time.NewTicker(2 * time.Second)
defer ticker.Stop()

lastStatus := j.Status

for {
select {
case <-ticker.C:
currentJob, err := repos.Job.Get(ctx, jobID)
if err != nil {
continue
}
if currentJob.Status != lastStatus {
data, _ := json.Marshal(currentJob)
if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
return
}
lastStatus = currentJob.Status
}
}
}
}
