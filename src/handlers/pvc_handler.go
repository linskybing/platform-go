package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/src/utils"
)

type PVCRequest struct {
	Namespace        string `form:"namespace" binding:"required"`
	Name             string `form:"name" binding:"required"`
	StorageClassName string `form:"storageClassName" binding:"required"`
	Size             string `form:"size" binding:"required"`
}

// @Summary query single PVC
// @Tags PVC
// @Produce json
// @Param namespace path string true "namespace"
// @Param name path string true "PVC name"
// @Success 200 {object} dto.PVC
// @Failure 404 {object} map[string]string
// @Router /pvc/{namespace}/{name} [get]
func GetPVCHandler(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")

	pvc, err := utils.GetPVC(ns, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pvc)
}

// @Summary List all PVC in Namespace
// @Tags PVC
// @Produce json
// @Param namespace path string true "namespace"
// @Success 200 {object} dto.PVC
// @Failure 500 {object} map[string]string
// @Router /pvc/list/{namespace} [get]
func ListPVCsHandler(c *gin.Context) {
	ns := c.Param("namespace")

	pvcs, err := utils.ListPVCs(ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pvcs)
}

// @Summary create PVC
// @Tags PVC
// @Accept x-www-form-urlencoded
// @Produce json
// @Security BearerAuth
// @Param pvc body PVCRequest true "PVC info"
// @Success 201 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /pvc [post]
func CreatePVCHandler(c *gin.Context) {
	var req PVCRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := utils.CreatePVC(req.Namespace, req.Name, req.StorageClassName, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "PVC created"})
}

// @Summary Expand PVC
// @Tags PVC
// @Accept x-www-form-urlencoded
// @Produce json
// @Security BearerAuth
// @Param pvc body PVCRequest true "PVC expand info"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /pvc/expand [put]
func ExpandPVCHandler(c *gin.Context) {
	var req PVCRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := utils.ExpandPVC(req.Namespace, req.Name, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PVC expanded"})
}

// @Summary Delete PVC
// @Tags PVC
// @Produce json
// @Security BearerAuth
// @Param namespace path string true "namespace"
// @Param name path string true "PVC name"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pvc/{namespace}/{name} [delete]
func DeletePVCHandler(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")

	err := utils.DeletePVC(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PVC deleted"})
}
