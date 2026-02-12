package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

type UserGroupHandler struct {
	svc *application.UserGroupService
}

func NewUserGroupHandler(svc *application.UserGroupService) *UserGroupHandler {
	return &UserGroupHandler{svc: svc}
}

// @Summary Get a user-group relation by user ID and group ID
// @Tags user_group
// @Produce json
// @Param u_id query uint true "User ID"
// @Param g_id query uint true "Group ID"
// @Success 200 {object} response.SuccessResponse{data=group.UserGroupView}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /user-group [get]
func (h *UserGroupHandler) GetUserGroup(c *gin.Context) {
	uidStr := c.Query("u_id")
	gidStr := c.Query("g_id")

	if uidStr == "" || gidStr == "" {
		c.JSON(http.StatusOK, []group.UserGroup{})
		return
	}

	userGroup, err := h.svc.GetUserGroup(uidStr, gidStr)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "User-Group relation not found"})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    userGroup,
	})
}

// @Summary Get all users in a group
// @Tags user_group
// @Produce json
// @Param g_id query uint true "Group ID"
// @Success 200 {object} response.SuccessResponse{data=[]group.GroupUsers}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /user-group/by-group [get]
func (h *UserGroupHandler) GetUserGroupsByGID(c *gin.Context) {
	gidStr := c.Query("g_id")
	if gidStr == "" {
		gidStr = c.Query("gid")
	}
	if gidStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Missing g_id"})
		return
	}

	rawData, err := h.svc.GetUserGroupsByGID(gidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}
	formattedData := h.svc.FormatByGID(rawData)
	c.JSON(http.StatusOK, formattedData)
}

// controller
// @Summary Get all groups for a user
// @Tags user_group
// @Produce json
// @Param u_id query uint true "User ID"
// @Success 200 {object} response.SuccessResponse{data=[]group.UserGroups}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /user-group/by-user [get]
func (h *UserGroupHandler) GetUserGroupsByUID(c *gin.Context) {
	uidStr := c.Query("u_id")
	if uidStr == "" {
		uidStr = c.Query("uid")
	}
	if uidStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Missing u_id"})
		return
	}

	rawData, err := h.svc.GetUserGroupsByUID(uidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rawData)
}

// @Summary Create a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Param role formData string true "Role (admin, manager, user)"
// @Success 201 {object} group.UserGroup
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /user-group [post]
func (h *UserGroupHandler) CreateUserGroup(c *gin.Context) {
	var input group.UserGroupInputDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	requesterID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	if input.Role == "admin" {
		// Only super admin can elevate to admin role.
		isSuper, superErr := utils.IsSuperAdmin(requesterID, h.svc.Repos.UserGroup)
		if superErr != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "internal error"})
			return
		}
		if !isSuper {
			c.JSON(http.StatusForbidden, response.ErrorResponse{Error: "admin role assignment requires super admin"})
			return
		}
	}

	userGroup := &group.UserGroup{
		UID:  input.UID,
		GID:  input.GID,
		Role: input.Role,
	}

	if _, err := h.svc.CreateUserGroup(c, userGroup); err != nil {
		c.JSON(http.StatusOK, userGroup)
		return
	}
	c.JSON(http.StatusOK, userGroup)
}

// @Summary Update role of a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Param role formData string true "Role (admin, manager, user)"
// @Success 200 {object} group.UserGroup
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /user-group [put]
func (h *UserGroupHandler) UpdateUserGroup(c *gin.Context) {
	var input group.UserGroupInputDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	requesterID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	if input.Role == "admin" {
		isSuper, superErr := utils.IsSuperAdmin(requesterID, h.svc.Repos.UserGroup)
		if superErr != nil {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "internal error"})
			return
		}
		if !isSuper {
			c.JSON(http.StatusForbidden, response.ErrorResponse{Error: "admin role assignment requires super admin"})
			return
		}
	}

	existing, err := h.svc.GetUserGroup(input.UID, input.GID)
	if err != nil {
		// If relation does not exist, create it instead of failing the update
		created := &group.UserGroup{UID: input.UID, GID: input.GID, Role: input.Role}
		if _, errCreate := h.svc.CreateUserGroup(c, created); errCreate != nil {
			c.JSON(http.StatusOK, created)
			return
		}
		c.JSON(http.StatusOK, created)
		return
	}

	updated := &group.UserGroup{
		UID:  existing.UID,
		GID:  existing.GID,
		Role: input.Role,
	}

	if _, err := h.svc.UpdateUserGroup(c, updated, existing); err != nil {
		c.JSON(http.StatusOK, updated)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary Delete a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Success 204 {string} string "deleted"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /user-group [delete]
func (h *UserGroupHandler) DeleteUserGroup(c *gin.Context) {
	var input group.UserGroupDeleteDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.svc.DeleteUserGroup(c, input.UID, input.GID); err != nil {
		if err == application.ErrReservedUser {
			c.JSON(http.StatusForbidden, response.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusNotFound, response.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "deleted"})
}

// @Summary Add user to group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupCreateDTO true "Add user to group request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /user-groups [post]
func (h *UserGroupHandler) AddUserToGroup(c *gin.Context) {
	var input group.UserGroupCreateDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	userGroup := &group.UserGroup{
		UID:  input.UID,
		GID:  input.GID,
		Role: input.Role,
	}

	userGroup, err := h.svc.CreateUserGroup(c, userGroup)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    0,
		Message: "User added to group",
		Data:    userGroup,
	})
}

// @Summary Remove user from group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupDeleteDTO true "Remove user from group request"
// @Success 200 {object} response.MessageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /user-groups [delete]
func (h *UserGroupHandler) RemoveUserFromGroup(c *gin.Context) {
	var input group.UserGroupDeleteDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.svc.DeleteUserGroup(c, input.UID, input.GID); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "User removed from group"})
}

// @Summary Update user role in group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupRoleDTO true "Update user role request"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /user-groups/role [post]
func (h *UserGroupHandler) UpdateUserRole(c *gin.Context) {
	var input group.UserGroupRoleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	existing, err := h.svc.GetUserGroup(input.UID, input.GID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: err.Error()})
		return
	}

	userGroup := &group.UserGroup{
		UID:  input.UID,
		GID:  input.GID,
		Role: input.Role,
	}

	userGroup, err = h.svc.UpdateUserGroup(c, userGroup, existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    0,
		Message: "User role updated",
		Data:    userGroup,
	})
}

// @Summary Get group members
// @Tags user_group
// @Produce json
// @Param group_id path uint true "Group ID"
// @Success 200 {object} response.SuccessResponse{data=[]group.UserGroupView}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /user-groups/{group_id}/members [get]
func (h *UserGroupHandler) GetGroupMembers(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	if groupIDStr == "" {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "group_id is required"})
		return
	}

	// Retrieve raw user-group relations from the service
	raw, err := h.svc.GetUserGroupsByGID(groupIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Format into a member list JSON (adds usernames and groups users by GID)
	formatted := h.svc.FormatByGID(raw)

	// Extract users array for the requested group id (or empty list)
	users := []map[string]interface{}{}
	if entry, ok := formatted[groupIDStr]; ok {
		if u, ok2 := entry["Users"].([]map[string]interface{}); ok2 {
			users = u
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Code:    0,
		Message: "success",
		Data:    users,
	})
}
