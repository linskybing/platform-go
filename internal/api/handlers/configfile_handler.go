package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
)

type ConfigFileHandler struct {
	svc *application.ConfigFileService
}

func NewConfigFileHandler(svc *application.ConfigFileService) *ConfigFileHandler {
	return &ConfigFileHandler{svc: svc}
}

// ListConfigFiles godoc
// @Summary List all config files
// @Tags config_files
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.ConfigFile
// @Failure 500 {object} response.ErrorResponse
// @Router /config-files [get]
func (h *ConfigFileHandler) ListConfigFilesHandler(c *gin.Context) {
	configFiles, err := h.svc.ListConfigFiles()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, configFiles, "Config files retrieved successfully")
}

// GetConfigFile godoc
// @Summary Get a config file by ID
// @Tags config_files
// @Security BearerAuth
// @Produce json
// @Param id path int true "Config File ID"
// @Success 200 {object} models.ConfigFile
// @Failure 400 {object} response.ErrorResponse "Invalid ID"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Router /config-files/{id} [get]
func (h *ConfigFileHandler) GetConfigFileHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config file ID")
		return
	}

	configFile, err := h.svc.GetConfigFile(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "config file not found")
		return
	}
	response.Success(c, configFile, "Config file retrieved successfully")
}

// CreateConfigFile godoc
// @Summary Create a new config file
// @Tags config_files
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param filename formData string true "Filename"
// @Param raw_yaml formData string true "Raw YAML content"
// @Param project_id formData int true "Project ID"
// @Success 201 {object} models.ConfigFile
// @Failure 400 {object} response.ErrorResponse "Bad request"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /config-files [post]
func (h *ConfigFileHandler) CreateConfigFileHandler(c *gin.Context) {
	var input configfile.CreateConfigFileInput
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid input: %v", err))
		return
	}

	if input.Filename == "" || input.RawYaml == "" || input.ProjectID == "" {
		response.Error(c, http.StatusBadRequest, "filename, raw_yaml, and project_id are required")
		return
	}

	claims, _ := c.MustGet("claims").(*types.Claims)
	configFile, err := h.svc.CreateConfigFile(c.Request.Context(), input, claims)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusCreated, configFile)
}

// UpdateConfigFile godoc
// @Summary Update a config file by ID
// @Tags config_files
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Config File ID"
// @Param filename formData string false "Filename"
// @Param raw_yaml formData string false "Raw YAML content"
// @Success 200 {object} models.ConfigFile
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /config-files/{id} [put]
func (h *ConfigFileHandler) UpdateConfigFileHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config file ID")
		return
	}

	var input configfile.ConfigFileUpdateDTO
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims, _ := c.MustGet("claims").(*types.Claims)
	updatedConfigFile, err := h.svc.UpdateConfigFile(c.Request.Context(), id, input, claims)
	if err != nil {
		if err == application.ErrConfigFileNotFound {
			response.Error(c, http.StatusNotFound, "config file not found")
		} else {
			response.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.Success(c, updatedConfigFile, "Config file updated successfully")
}

// DeleteConfigFile godoc
// @Summary Delete a config file
// @Tags config_files
// @Security BearerAuth
// @Param id path int true "ConfigFile ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /config-files/{id} [delete]
func (h *ConfigFileHandler) DeleteConfigFileHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config file ID")
		return
	}

	claims, _ := c.MustGet("claims").(*types.Claims)
	err = h.svc.DeleteConfigFile(c.Request.Context(), id, claims)
	if err != nil {
		if err == application.ErrConfigFileNotFound {
			response.Error(c, http.StatusNotFound, "config file not found")
		} else {
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListConfigFilesByProjectID godoc
// @Summary List config files by project ID
// @Tags config_files
// @Security BearerAuth
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {array} models.ConfigFile
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /projects/{id}/config-files [get]
func (h *ConfigFileHandler) ListConfigFilesByProjectIDHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid project_id")
		return
	}

	configFiles, err := h.svc.ListConfigFilesByProjectID(id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, configFiles, "Config files retrieved successfully")
}

// CreateInstanceHandler godoc
// @Summary Instantiate a config file instance
// @Description Creates a Kubernetes instance from a config file.
// @Tags Instance
// @Security BearerAuth
// @Produce json
// @Param id path int true "Config File ID"
// @Success 200 {object} response.MessageResponse "Instance created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid config file ID or validation error"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /instance/{id} [post]
func (h *ConfigFileHandler) CreateInstanceHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config id")
		return
	}
	claims, _ := c.MustGet("claims").(*types.Claims)
	err = h.svc.CreateInstance(c.Request.Context(), id, claims)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil, "create successfully")
}

// Destruce ConfigFile Instance godoc
// @Summary Destruct a config file instance
// @Tags Instance
// @Security BearerAuth
// @Produce json
// @Param id path int true "Config File ID"
// @Success 204 "No content"
// @Failure 400 {object} response.ErrorResponse "Invalid ID"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /instance/{id} [delete]
func (h *ConfigFileHandler) DestructInstanceHandler(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid config id")
		return
	}
	claims, _ := c.MustGet("claims").(*types.Claims)
	err = h.svc.DeleteInstance(c.Request.Context(), id, claims)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
