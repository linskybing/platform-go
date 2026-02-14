package user

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
// @Success 200 {object} response.StandardResponse{data=group.UserGroupView}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Router /user-group [get]
func (h *UserGroupHandler) GetUserGroup(c *gin.Context) {
	uidStr := c.Query("u_id")
	gidStr := c.Query("g_id")

	if uidStr == "" || gidStr == "" {
		response.Success(c, []group.UserGroup{}, "No user-group relations found")
		return
	}

	userGroup, err := h.svc.GetUserGroup(uidStr, gidStr)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User-Group relation not found")
		return
	}

	response.Success(c, userGroup, "User-Group relation retrieved successfully")
}

// @Summary Get all users in a group
// @Tags user_group
// @Produce json
// @Param g_id query uint true "Group ID"
// @Success 200 {object} response.StandardResponse{data=[]group.GroupUsers}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /user-group/by-group [get]
func (h *UserGroupHandler) GetUserGroupsByGID(c *gin.Context) {
	gidStr := c.Query("g_id")
	if gidStr == "" {
		gidStr = c.Query("gid")
	}
	if gidStr == "" {
		response.Error(c, http.StatusBadRequest, "Missing g_id")
		return
	}

	rawData, err := h.svc.GetUserGroupsByGID(gidStr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	formattedData := h.svc.FormatByGID(rawData)
	response.Success(c, formattedData, "Users in group retrieved successfully")
}

// controller
// @Summary Get all groups for a user
// @Tags user_group
// @Produce json
// @Param u_id query uint true "User ID"
// @Success 200 {object} response.StandardResponse{data=[]group.UserGroups}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /user-group/by-user [get]
func (h *UserGroupHandler) GetUserGroupsByUID(c *gin.Context) {
	uidStr := c.Query("u_id")
	if uidStr == "" {
		uidStr = c.Query("uid")
	}
	if uidStr == "" {
		response.Error(c, http.StatusBadRequest, "Missing u_id")
		return
	}

	rawData, err := h.svc.GetUserGroupsByUID(uidStr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, rawData, "Groups for user retrieved successfully")
}

// @Summary Create a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Param role formData string true "Role (admin, manager, user)"
// @Success 201 {object} response.StandardResponse{data=group.UserGroup}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 401 {object} response.StandardResponse{data=nil}
// @Failure 403 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /user-group [post]
func (h *UserGroupHandler) CreateUserGroup(c *gin.Context) {
	var input group.UserGroupInputDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	requesterID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if input.Role == "admin" {
		// Only super admin can elevate to admin role.
		isSuper, superErr := utils.IsSuperAdmin(requesterID, h.svc.Repos.UserGroup)
		if superErr != nil {
			response.Error(c, http.StatusInternalServerError, "Internal error")
			return
		}
		if !isSuper {
			response.Error(c, http.StatusForbidden, "Admin role assignment requires super admin")
			return
		}
	}

	userGroup := &group.UserGroup{
		UserID:  input.UID,
		GroupID: input.GID,
		Role:    input.Role,
	}

	if _, err := h.svc.CreateUserGroup(c, userGroup); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error()) // assuming internal error if creation fails
		return
	}
	response.Success(c, userGroup, "User-Group relation created successfully")
}

// @Summary Update role of a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Param role formData string true "Role (admin, manager, user)"
// @Success 200 {object} response.StandardResponse{data=group.UserGroup}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 401 {object} response.StandardResponse{data=nil}
// @Failure 403 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /user-group [put]
func (h *UserGroupHandler) UpdateUserGroup(c *gin.Context) {
	var input group.UserGroupInputDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	requesterID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	if input.Role == "admin" {
		isSuper, superErr := utils.IsSuperAdmin(requesterID, h.svc.Repos.UserGroup)
		if superErr != nil {
			response.Error(c, http.StatusInternalServerError, "Internal error")
			return
		}
		if !isSuper {
			response.Error(c, http.StatusForbidden, "Admin role assignment requires super admin")
			return
		}
	}

	existing, err := h.svc.GetUserGroup(input.UID, input.GID)
	if err != nil {
		// If relation does not exist, create it instead of failing the update
		created := &group.UserGroup{UserID: input.UID, GroupID: input.GID, Role: input.Role}
		if _, errCreate := h.svc.CreateUserGroup(c, created); errCreate != nil {
			response.Error(c, http.StatusInternalServerError, errCreate.Error())
			return
		}
		response.Success(c, created, "User-Group relation created/updated successfully")
		return
	}

	updated := &group.UserGroup{
		UserID:  existing.UserID,
		GroupID: existing.GroupID,
		Role:    input.Role,
	}

	if _, err := h.svc.UpdateUserGroup(c, updated, existing); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, updated, "User-Group relation updated successfully")
}

// @Summary Delete a user-group relation
// @Tags user_group
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param u_id formData uint true "User ID"
// @Param g_id formData uint true "Group ID"
// @Success 200 {object} response.StandardResponse{data=nil} "User-Group relation deleted successfully"
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 403 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /user-group [delete]
func (h *UserGroupHandler) DeleteUserGroup(c *gin.Context) {
	var input group.UserGroupDeleteDTO
	if err := c.ShouldBindQuery(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteUserGroup(c, input.UID, input.GID); err != nil {
		if err == application.ErrReservedUser {
			response.Error(c, http.StatusForbidden, err.Error())
		} else {
			response.Error(c, http.StatusNotFound, err.Error()) // Assuming 404 if not found
		}
		return
	}

	response.Success(c, nil, "User-Group relation deleted successfully")
}

// @Summary Add user to group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupCreateDTO true "Add user to group request"
// @Success 200 {object} response.StandardResponse{data=group.UserGroup}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 409 {object} response.StandardResponse{data=nil}
// @Router /user-groups [post]
func (h *UserGroupHandler) AddUserToGroup(c *gin.Context) {
	var input group.UserGroupCreateDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userGroup := &group.UserGroup{
		UserID:  input.UID,
		GroupID: input.GID,
		Role:    input.Role,
	}

	userGroup, err := h.svc.CreateUserGroup(c, userGroup)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error()) // Assuming bad request if creation fails for some reason
		return
	}

	response.Success(c, userGroup, "User added to group successfully")
}

// @Summary Remove user from group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupDeleteDTO true "Remove user from group request"
// @Success 200 {object} response.StandardResponse{data=nil}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 403 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Router /user-groups [delete]
func (h *UserGroupHandler) RemoveUserFromGroup(c *gin.Context) {
	var input group.UserGroupDeleteDTO
	if err := c.ShouldBindQuery(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.DeleteUserGroup(c, input.UID, input.GID); err != nil {
		if err == application.ErrReservedUser {
			response.Error(c, http.StatusForbidden, err.Error())
		} else {
			response.Error(c, http.StatusNotFound, err.Error())
		}
		return
	}

	response.Success(c, nil, "User removed from group successfully")
}

// @Summary Update user role in group
// @Tags user_group
// @Accept json
// @Produce json
// @Param request body group.UserGroupRoleDTO true "Update user role request"
// @Success 200 {object} response.StandardResponse{data=group.UserGroup}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Router /user-groups/role [post]
func (h *UserGroupHandler) UpdateUserRole(c *gin.Context) {
	var input group.UserGroupRoleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := h.svc.GetUserGroup(input.UID, input.GID)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	userGroup := &group.UserGroup{
		UserID:  input.UID,
		GroupID: input.GID,
		Role:    input.Role,
	}

	userGroup, err = h.svc.UpdateUserGroup(c, userGroup, existing)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, userGroup, "User role updated successfully")
}

// @Summary Get group members
// @Tags user_group
// @Produce json
// @Param group_id path uint true "Group ID"
// @Success 200 {object} response.StandardResponse{data=[]group.UserGroupView}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 404 {object} response.StandardResponse{data=nil}
// @Router /user-groups/{group_id}/members [get]
func (h *UserGroupHandler) GetGroupMembers(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	if groupIDStr == "" {
		response.Error(c, http.StatusBadRequest, "Group ID is required")
		return
	}

	// Retrieve raw user-group relations from the service
	raw, err := h.svc.GetUserGroupsByGID(groupIDStr)
	if err != nil {
		response.Error(c, http.StatusNotFound, err.Error())
		return
	}

	// Format into a member list JSON (adds usernames and groups users by GID)
	formatted := h.svc.FormatByGID(raw)

	// Extract users array for the requested group id (or empty list)
	users := []map[string]interface{}{}
	if entry, ok := formatted[groupIDStr]; ok {
		if u, ok2 := entry["users"].([]map[string]interface{}); ok2 {
			users = u
		}
	}

	response.Success(c, users, "Group members retrieved successfully")
}
