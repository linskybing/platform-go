package group

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	domaingroup "github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/gorm"
)

type GroupHandler struct {
	svc *application.GroupService
}

func NewGroupHandler(svc *application.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

// GetGroups godoc
// @Summary List all groups
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]domaingroup.Group}
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /groups [get]
func (h *GroupHandler) GetGroups(c *gin.Context) {
	groups, err := h.svc.ListGroups()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, groups, "Groups retrieved successfully")
}

// GetGroupByID godoc
// @Summary Get group by ID
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {object} response.StandardResponse{data=domaingroup.Group}
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid group id"
// @Failure 404 {object} response.StandardResponse{data=nil} "Group not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /groups/{id} [get]
func (h *GroupHandler) GetGroupByID(c *gin.Context) {
	gid, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid group ID")
		return
	}
	group, err := h.svc.GetGroup(gid)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Group not found")
		return
	}

	response.Success(c, group, "Group retrieved successfully")
}

// CreateGroup godoc
// @Summary Create a new group
// @Tags groups
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param group_name formData string true "Group name"
// @Param description formData string false "Description"
// @Success 201 {object} response.StandardResponse{data=domaingroup.Group}
// @Failure 400 {object} response.StandardResponse{data=nil} "Bad request"
// @Failure 403 {object} response.StandardResponse{data=nil} "Forbidden (reserved name)"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /groups [post]
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var input domaingroup.GroupCreateDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	group, err := h.svc.CreateGroup(c, input)
	if err != nil {
		if err == application.ErrReservedGroupName {
			response.Error(c, http.StatusForbidden, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, group, "Group created successfully")
}

// UpdateGroup godoc
// @Summary Update group by ID
// @Tags groups
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Group ID"
// @Param group_name formData string false "Group name"
// @Param description formData string false "Description"
// @Success 200 {object} models.Group
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 404 {object} response.ErrorResponse "Group not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /groups/{id} [put]
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid group id"})
		return
	}

	var input domaingroup.GroupUpdateDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	group, err := h.svc.UpdateGroup(c, id, input)
	if err != nil {
		if err == application.ErrReservedGroupName {
			c.JSON(http.StatusForbidden, response.ErrorResponse{Error: err.Error()})
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "group not found"})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup godoc
// @Summary Delete group by ID
// @Tags groups
// @Security BearerAuth
// @Produce json
// @Param id path int true "Group ID"
// @Success 200 {object} response.MessageResponse "Group deleted"
// @Failure 400 {object} response.ErrorResponse "Invalid group id"
// @Failure 403 {object} response.ErrorResponse "Forbidden to delete 'super' group"
// @Failure 404 {object} response.ErrorResponse "Group not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /groups/{id} [delete]
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid group id"})
		return
	}

	err = h.svc.DeleteGroup(c, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, response.ErrorResponse{Error: "group not found"})
		} else if err == application.ErrReservedGroupName {
			c.JSON(http.StatusForbidden, response.ErrorResponse{Error: "super group can't be removed"})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response.MessageResponse{Message: "Group deleted"})
}
