package form

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	domainform "github.com/linskybing/platform-go/internal/domain/form"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

type FormHandler struct {
	service *application.FormService
}

func NewFormHandler(service *application.FormService) *FormHandler {
	return &FormHandler{service: service}
}

// CreateForm godoc
// @Summary Create a new form
// @Description Create a new form for a project (project_id optional)
// @Tags forms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body form.CreateFormDTO true "Create form request"
// @Success 200 {object} form.Form
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms [post]
func (h *FormHandler) CreateForm(c *gin.Context) {
	var input domainform.CreateFormDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	form, err := h.service.CreateForm(userID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// GetMyForms godoc
// @Summary List current user's forms
// @Tags forms
// @Security BearerAuth
// @Produce json
// @Success 200 {array} form.Form
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms/my [get]
func (h *FormHandler) GetMyForms(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	forms, err := h.service.GetUserForms(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, forms)
}

// GetAllForms godoc
// @Summary List accessible forms
// @Tags forms
// @Security BearerAuth
// @Produce json
// @Success 200 {array} form.Form
// @Failure 500 {object} response.ErrorResponse
// @Router /forms [get]
func (h *FormHandler) GetAllForms(c *gin.Context) {
	forms, err := h.service.GetAllForms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, forms)
}

// UpdateFormStatus godoc
// @Summary Update form status
// @Tags forms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param request body form.UpdateFormStatusDTO true "Update status request"
// @Success 200 {object} form.Form
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms/{id}/status [put]
func (h *FormHandler) UpdateFormStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Invalid ID"})
		return
	}

	var input domainform.UpdateFormStatusDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	form, err := h.service.UpdateFormStatus(fmt.Sprintf("%d", id), input.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, form)
}

// CreateMessage godoc
// @Summary Add a message to a form
// @Tags forms
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param request body form.CreateFormMessageDTO true "Create message request"
// @Success 200 {object} form.FormMessage
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms/{id}/messages [post]
func (h *FormHandler) CreateMessage(c *gin.Context) {
	formIDStr := c.Param("id")
	formID, err := strconv.ParseUint(formIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Invalid ID"})
		return
	}

	var input domainform.CreateFormMessageDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Unauthorized"})
		return
	}

	msg, err := h.service.AddMessage(fmt.Sprintf("%d", formID), userID, input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}

// ListMessages godoc
// @Summary List messages for a form
// @Tags forms
// @Security BearerAuth
// @Produce json
// @Param id path int true "Form ID"
// @Success 200 {array} form.FormMessage
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /forms/{id}/messages [get]
func (h *FormHandler) ListMessages(c *gin.Context) {
	formIDStr := c.Param("id")
	_, err := strconv.ParseUint(formIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Invalid ID"})
		return
	}

	msgs, err := h.service.ListMessages(formIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}
