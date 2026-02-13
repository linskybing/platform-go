package cluster

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/cluster"
	"github.com/linskybing/platform-go/pkg/response"
)

type ClusterHandler struct {
	svc *cluster.ClusterService
}

func NewClusterHandler(svc *cluster.ClusterService) *ClusterHandler {
	return &ClusterHandler{svc: svc}
}

// GetClusterSummary godoc
// @Summary Get cluster resource summary
// @Tags cluster
// @Security BearerAuth
// @Produce json
// @Success 200 {object} cluster.ClusterSummary
// @Failure 500 {object} response.ErrorResponse
// @Router /cluster/summary [get]
func (h *ClusterHandler) GetClusterSummary(c *gin.Context) {
	summary, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, summary)
}

// ListClusterNodes godoc
// @Summary List cluster nodes with resources
// @Tags cluster
// @Security BearerAuth
// @Produce json
// @Success 200 {array} cluster.NodeResourceInfo
// @Failure 500 {object} response.ErrorResponse
// @Router /cluster/nodes [get]
func (h *ClusterHandler) ListClusterNodes(c *gin.Context) {
	summary, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, summary.Nodes)
}

// GetClusterNode godoc
// @Summary Get cluster node resource detail
// @Tags cluster
// @Security BearerAuth
// @Produce json
// @Param name path string true "Node name"
// @Success 200 {object} cluster.NodeResourceInfo
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /cluster/nodes/{name} [get]
func (h *ClusterHandler) GetClusterNode(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, http.StatusBadRequest, "node name required")
		return
	}

	summary, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	for _, node := range summary.Nodes {
		if node.Name == name {
			c.JSON(http.StatusOK, node)
			return
		}
	}
	response.Error(c, http.StatusNotFound, "node not found")
}

// ListPodGPUUsage godoc
// @Summary List per-pod GPU usage
// @Tags cluster
// @Security BearerAuth
// @Produce json
// @Success 200 {array} cluster.PodGPUUsage
// @Failure 500 {object} response.ErrorResponse
// @Router /cluster/gpu-usage [get]
func (h *ClusterHandler) ListPodGPUUsage(c *gin.Context) {
	summary, err := h.svc.GetSummary(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, summary.PodGPUUsages)
}
