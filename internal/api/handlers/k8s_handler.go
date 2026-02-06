package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	appk8s "github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/pkg/filebrowser"
	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

type K8sHandler struct {
	K8sService     *appk8s.K8sService
	UserService    *application.UserService
	ProjectService *application.ProjectService
}

func NewK8sHandler(K8sService *appk8s.K8sService, UserService *application.UserService, ProjectService *application.ProjectService) *K8sHandler {
	return &K8sHandler{
		K8sService:     K8sService,
		UserService:    UserService,
		ProjectService: ProjectService,
	}
}

// GetPodLogs godoc
// @Summary Get Pod Logs
// @Tags k8s
// @Produce plain
// @Param ns path string true "Namespace"
// @Param name path string true "Pod name"
// @Param container query string false "Container name"
// @Param follow query bool false "Follow logs"
// @Param tailLines query int false "Tail lines"
// @Router /k8s/namespaces/{ns}/pods/{name}/logs [get]
func (h *K8sHandler) GetPodLogs(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	container := c.Query("container")
	follow := strings.ToLower(c.Query("follow")) == "true"
	var tailLinesPtr *int64
	if raw := c.Query("tailLines"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			tailLinesPtr = &v
		}
	}

	if k8s.Clientset == nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "kubernetes client not configured"})
		return
	}

	req := k8s.Clientset.CoreV1().Pods(ns).GetLogs(name, &corev1.PodLogOptions{
		Container: container,
		Follow:    follow,
		TailLines: tailLinesPtr,
	})

	stream, err := req.Stream(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}
	defer func() { _ = stream.Close() }()

	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Status(http.StatusOK)
	flusher, _ := c.Writer.(http.Flusher)
	buf := make([]byte, 8192)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			_, _ = c.Writer.Write(buf[:n])
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			if err == io.EOF {
				return
			}
			return
		}
	}
}

// OpenMyDrive godoc
// @Summary Open user's global file browser
// @Description Spins up a temporary FileBrowser pod connected to the user's storage hub (NFS Client).
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Returns nodePort"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /k8s/user-storage/browse [post]
func (h *K8sHandler) OpenMyDrive(c *gin.Context) {
	claimsVal, _ := c.Get("claims")
	claims := claimsVal.(*types.Claims)
	userID := claims.UserID
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, err := h.UserService.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "User not found: " + err.Error()})
		return
	}

	_, err = h.K8sService.OpenUserGlobalFileBrowser(c, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "Failed to start file browser: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User file browser ready",
	})
}

// StopMyDrive godoc
// @Summary Stop user's global file browser
// @Description Terminates the temporary FileBrowser pod and service for the user's storage hub.
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.MessageResponse "Resources cleaned up"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /k8s/user-storage/browse [delete]
func (h *K8sHandler) StopMyDrive(c *gin.Context) {
	claimsVal, _ := c.Get("claims")
	claims := claimsVal.(*types.Claims)
	userID := claims.UserID
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	user, err := h.UserService.FindUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "User not found: " + err.Error()})
		return
	}

	err = h.K8sService.StopUserGlobalFileBrowser(c, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "Failed to stop file browser: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{
		Message: "User file browser stopped successfully",
	})
}

// UserStorageProxy handles all traffic to the FileBrowser
// @Summary Proxy to user file browser
// @Tags k8s
// @Security BearerAuth
// @Router /k8s/user-storage/proxy/*path [all]
func (h *K8sHandler) UserStorageProxy(c *gin.Context) {
	claimsVal, _ := c.Get("claims")
	claims := claimsVal.(*types.Claims)
	userID := claims.UserID
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, _ := h.UserService.FindUserByID(userID)
	safeUsername := strings.ToLower(user.Username)

	// Construct service name and namespace based on naming convention
	serviceName := fmt.Sprintf("fb-hub-svc-%s", safeUsername)
	namespace := fmt.Sprintf("user-%s-storage", safeUsername)

	// Use shared proxy handler from filebrowser package
	proxyHandler := filebrowser.ProxyHandler(filebrowser.ProxyConfig{
		ServiceName: serviceName,
		Namespace:   namespace,
		PathPrefix:  "/k8s/user-storage/proxy",
	})

	proxyHandler(c)
}
