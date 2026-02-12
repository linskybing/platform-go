package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application/k8s"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// GroupStorageHandler handles HTTP endpoints for group storages (non-breaking aliases)
type GroupStorageHandler struct {
	storageMgr  *k8s.StorageManager
	fbManager   *k8s.FileBrowserManager
	permManager *k8s.PermissionManager
	auditRepo   repository.AuditRepo
}

func NewGroupStorageHandler(sm *k8s.StorageManager, fb *k8s.FileBrowserManager, pm *k8s.PermissionManager, auditRepo repository.AuditRepo) *GroupStorageHandler {
	return &GroupStorageHandler{storageMgr: sm, fbManager: fb, permManager: pm, auditRepo: auditRepo}
}

// ListGroupStorages godoc
// @Summary List storages for a group
// @Tags Group Storage
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {array} storage.GroupPVCSpec
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/group/{id} [get]
func (h *GroupStorageHandler) ListGroupStorages(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "Invalid group id")
		return
	}
	// allow both numeric and string ids; backend storage manager expects string gid
	groupID := id
	pvcs, err := h.storageMgr.ListGroupPVCs(c.Request.Context(), groupID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list group storages: "+err.Error())
		return
	}
	response.Success(c, pvcs, "Listed group storages")
}

// GetMyGroupStorages godoc
// @Summary List current user's accessible storages
// @Tags Group Storage
// @Security BearerAuth
// @Produce json
// @Success 200 {array} storage.GroupPVCWithPermissions
// @Failure 500 {object} response.Response
// @Router /storage/my-storages [get]
func (h *GroupStorageHandler) GetMyGroupStorages(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	perms := h.permManager
	if perms == nil {
		response.Success(c, []storage.GroupPVCWithPermissions{}, "")
		return
	}

	// Fetch user's permissions and map to PVC info
	userPerms, err := perms.ListUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list user permissions: "+err.Error())
		return
	}

	// groupID -> []pvcIDs
	groups := make(map[string][]string)
	for _, p := range userPerms {
		groups[p.GroupID] = append(groups[p.GroupID], p.PVCID)
	}

	var result []storage.GroupPVCWithPermissions
	for gid, pvcIDs := range groups {
		pvcs, _ := h.storageMgr.ListGroupPVCs(c.Request.Context(), gid)
		// index pvcs by id or pvc-uuid
		for _, pv := range pvcs {
			for _, want := range pvcIDs {
				if pv.ID == want || pv.PVCName == want {
					// find permission record
					var up storage.GroupStoragePermission
					for _, p := range userPerms {
						if p.GroupID == gid && (p.PVCID == want) {
							up = p
							break
						}
					}
					result = append(result, storage.GroupPVCWithPermissions{
						GroupPVCSpec: storage.GroupPVCSpec{
							ID:           pv.ID,
							GroupID:      pv.GroupID,
							Name:         pv.Name,
							PVCName:      pv.PVCName,
							Capacity:     pv.Capacity,
							StorageClass: pv.StorageClass,
							Status:       pv.Status,
							AccessMode:   pv.AccessMode,
							CreatedAt:    pv.CreatedAt,
							CreatedBy:    pv.CreatedBy,
						},
						UserPermission: up.Permission,
						CanAccess:      up.Permission != storage.PermissionNone,
						CanModify:      up.Permission == storage.PermissionReadWrite,
					})
				}
			}
		}
	}

	response.Success(c, result, "")
}

// CreateGroupStorage godoc
// @Summary Create a new group PVC
// @Tags Group Storage
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Group ID"
// @Param request body storage.CreateGroupStorageRequest true "Create request"
// @Success 200 {object} storage.GroupPVCSpec
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/{id}/storage [post]
func (h *GroupStorageHandler) CreateGroupStorage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, http.StatusBadRequest, "Invalid group id")
		return
	}
	var req storage.CreateGroupStorageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}
	userID, _ := utils.GetUserIDFromContext(c)
	pvc, err := h.storageMgr.CreateGroupPVC(c.Request.Context(), id, &req, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create group storage: "+err.Error())
		return
	}

	// Audit log
	utils.LogAuditWithConsole(c, "create", "group_storage", pvc.ID, nil, pvc, "Created group storage", h.auditRepo)

	response.Success(c, pvc, "Group storage created")
}

// DeleteGroupStorage godoc
// @Summary Delete a group PVC
// @Tags Group Storage
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Param pvcId path string true "PVC ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/{id}/storage/{pvcId} [delete]
func (h *GroupStorageHandler) DeleteGroupStorage(c *gin.Context) {
	id := c.Param("id")
	pvcId := c.Param("pvcId")
	if id == "" || pvcId == "" {
		response.Error(c, http.StatusBadRequest, "Invalid parameters")
		return
	}
	if err := h.storageMgr.DeleteGroupPVC(c.Request.Context(), pvcId); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete pvc: "+err.Error())
		return
	}

	// Audit log
	utils.LogAuditWithConsole(c, "delete", "group_storage", pvcId, gin.H{"group_id": id, "pvc_id": pvcId}, nil, "Deleted group storage", h.auditRepo)

	response.Success(c, nil, "Deleted")
}

// StartFileBrowser godoc
// @Summary Start FileBrowser for a group PVC
// @Tags Group Storage
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Param pvcId path string true "PVC ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/{id}/storage/{pvcId}/start [post]
func (h *GroupStorageHandler) StartFileBrowser(c *gin.Context) {
	id := c.Param("id")
	pvcId := c.Param("pvcId")
	if id == "" || pvcId == "" {
		response.Error(c, http.StatusBadRequest, "Invalid parameters")
		return
	}
	// request FileBrowserManager to start (it requires lower-level info)
	// Build a FileBrowserAccessRequest and call GetFileBrowserAccess
	var req storage.FileBrowserAccessRequest
	req.GroupID = id
	req.PVCID = pvcId
	userID, err := utils.GetUserIDFromContext(c)
	if err == nil {
		req.UserID = userID
	}
	resp, err := h.fbManager.GetFileBrowserAccess(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to start filebrowser: "+err.Error())
		return
	}
	response.Success(c, resp, "Filebrowser started")
}

// StopFileBrowser godoc
// @Summary Stop FileBrowser for a group PVC
// @Tags Group Storage
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Param pvcId path string true "PVC ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /storage/{id}/storage/{pvcId}/stop [delete]
func (h *GroupStorageHandler) StopFileBrowser(c *gin.Context) {
	id := c.Param("id")
	pvcId := c.Param("pvcId")
	if id == "" || pvcId == "" {
		response.Error(c, http.StatusBadRequest, "Invalid parameters")
		return
	}
	userID, _ := utils.GetUserIDFromContext(c)
	if err := h.fbManager.StopFileBrowser(c.Request.Context(), id, pvcId, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to stop filebrowser: "+err.Error())
		return
	}
	response.Success(c, nil, "Stop requested")
}
