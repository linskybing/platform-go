package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWatchJobStatusHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobRepo := mock.NewMockJobRepo(ctrl)
	repos := &repository.Repos{Job: mockJobRepo}

	testJob := &job.Job{ID: "test-job-1", Status: "running"}

	// Called initially, then via ticker
	mockJobRepo.EXPECT().Get(gomock.Any(), "test-job-1").Return(testJob, nil).AnyTimes()

	r := gin.New()
	r.GET("/ws/jobs/:id", func(c *gin.Context) {
		WatchJobStatusHandler(c, repos)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/jobs/test-job-1"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	requireNoError(t, err)
	defer ws.Close()

	_, msg, err := ws.ReadMessage()
	requireNoError(t, err)

	var receivedJob job.Job
	err = json.Unmarshal(msg, &receivedJob)
	requireNoError(t, err)
	assert.Equal(t, "test-job-1", receivedJob.ID)

	// Keep connection open briefly to trigger reader loop
	time.Sleep(10 * time.Millisecond)
}

func TestWatchJobStatusHandler_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repos := &repository.Repos{}

	r := gin.New()
	r.GET("/ws/jobs", func(c *gin.Context) {
		WatchJobStatusHandler(c, repos)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ws/jobs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWatchJobStatusHandler_JobNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobRepo := mock.NewMockJobRepo(ctrl)
	repos := &repository.Repos{Job: mockJobRepo}

	mockJobRepo.EXPECT().Get(gomock.Any(), "test-job-1").Return(nil, assert.AnError)

	r := gin.New()
	r.GET("/ws/jobs/:id", func(c *gin.Context) {
		WatchJobStatusHandler(c, repos)
	})
	ts := httptest.NewServer(r)
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/jobs/test-job-1"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	requireNoError(t, err)
	defer ws.Close()

	_, msg, err := ws.ReadMessage()
	requireNoError(t, err)
	assert.Contains(t, string(msg), "job not found")
}

func requireNoError(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatalf("unexpected error: %v", err)
	}
}

func BenchmarkJobJSONMarshal(b *testing.B) {
	j := &job.Job{
		ID:          "bench-job",
		Status:      "running",
		ProjectID:   "project-abc",
		UserID:      "user-123",
		SubmittedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(j)
	}
}
