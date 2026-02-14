package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/linskybing/platform-go/internal/application"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
	"gorm.io/gorm"
)

type UserHandler struct {
	svc *application.UserService
}

func NewUserHandler(svc *application.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register godoc
// @Summary User registration
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param input body user.CreateUserInput true "User registration info"
// @Success 201 {object} response.StandardResponse{data=nil} "User registered successfully"
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid input"
// @Failure 409 {object} response.StandardResponse{data=nil} "Username already taken"
// @Failure 500 {object} response.StandardResponse{data=nil} "Failed to create user"
// @Router /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var input user.CreateUserInput

	if err := c.ShouldBind(&input); err != nil {
		// Try to produce friendly validation messages for the frontend
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			msgs := make([]string, 0, len(verr))

			labels := map[string]string{
				"Username": "username",
				"Password": "password",
				"Email":    "email",
				"FullName": "full name",
				"Type":     "type",
				"Status":   "status",
			}

			for _, fe := range verr {
				field := fe.StructField()
				lbl, ok := labels[field]
				if !ok {
					lbl = strings.ToLower(field)
				}

				var msg string
				switch fe.Tag() {
				case "required":
					msg = fmt.Sprintf("%s is required", lbl)
				case "min":
					msg = fmt.Sprintf("%s must be at least %s characters", lbl, fe.Param())
				case "max":
					msg = fmt.Sprintf("%s must be at most %s characters", lbl, fe.Param())
				case "email":
					msg = fmt.Sprintf("%s must be a valid email address", lbl)
				case "oneof":
					msg = fmt.Sprintf("%s must be one of [%s]", lbl, fe.Param())
				default:
					msg = fmt.Sprintf("%s is invalid", lbl)
				}
				msgs = append(msgs, msg)
			}

			response.Error(c, http.StatusBadRequest, strings.Join(msgs, "; "))
			return
		}

		response.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	err := h.svc.RegisterUser(input)
	if err != nil {
		if errors.Is(err, application.ErrUsernameTaken) {
			response.Error(c, http.StatusConflict, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, nil, "User registered successfully")
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} response.StandardResponse{data=response.TokenData} "JWT token and user info"
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid input"
// @Failure 401 {object} response.StandardResponse{data=nil} "Invalid username or password"
// @Failure 500 {object} response.StandardResponse{data=nil} "Failed to generate token"
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	user, token, isAdmin, err := h.svc.LoginUser(req.Username, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"token",
		token,
		3600,
		"/",
		"",
		config.IsProduction, // Secure only in production
		true,
	)

	// Return token in both cookie and response body for flexibility
	response.Success(c, response.TokenData{
		Token:    token,
		UID:      user.ID,
		Username: user.Username,
		IsAdmin:  isAdmin,
	}, "Login successful")
}

// Logout godoc
// @Summary User logout
// @Tags auth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=nil} "Logout successful"
// @Router /logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	response.Success(c, nil, "Logout successful")
}

// AuthStatus godoc
// @Summary Check auth token status
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=object}
// @Failure 401 {object} response.StandardResponse{data=nil}
// @Router /auth/status [get]
func (h *UserHandler) AuthStatus(c *gin.Context) {
	uid, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, "token expired")
		return
	}
	response.Success(c, map[string]interface{}{
		"status":  "valid",
		"user_id": uid,
	}, "Token is valid")
}

// GetUsers godoc
// @Summary List all users
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]user.UserWithSuperAdmin}
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.svc.ListUsers()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Success(c, []user.UserWithSuperAdmin{}, "No users found") // Return empty array for no records
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, users, "Users retrieved successfully")
}

// ListUsersPaging godoc
// @Summary List all users with pagination
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Success 200 {object} response.StandardResponse{data=[]user.UserWithSuperAdmin}
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /users/paging [get]
func (h *UserHandler) ListUsersPaging(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	users, err := h.svc.ListUserByPaging(page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, users, "Paginated users retrieved successfully")
}

// GetUserByID godoc
// @Summary Get user by ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.StandardResponse{data=user.UserWithSuperAdmin}
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid user id"
// @Failure 404 {object} response.StandardResponse{data=nil} "User not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.svc.FindUserByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}
	response.Success(c, user, "User retrieved successfully")
}

// UpdateUser updates the information of a user by ID.
// @Summary Update user
// @Security BearerAuth
// @Description Partially update user's email, full name, type, status, or password.
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "User ID"
// @Param old_password formData string false "Old password (required if updating password)"
// @Param password formData string false "New password"
// @Param email formData string false "Email"
// @Param full_name formData string false "Full name"
// @Param type formData string false "User type: origin or oauth2"
// @Param status formData string false "User status: online, offline, delete"
// @Success 200 {object} response.StandardResponse{data=user.UserDTO} "Updated user info"
// @Failure 400 {object} response.StandardResponse{data=nil} "Bad request error"
// @Failure 401 {object} response.StandardResponse{data=nil} "Unauthorized"
// @Failure 404 {object} response.StandardResponse{data=nil} "User not found"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var input user.UpdateUserInput
	if err := c.ShouldBind(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	updatedUser, err := h.svc.UpdateUser(id, input)
	if err != nil {
		switch err {
		case application.ErrUserNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case application.ErrMissingOldPassword, application.ErrIncorrectPassword:
			response.Error(c, http.StatusUnauthorized, err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, updatedUser, "User updated successfully")
}

// DeleteUser godoc
// @Summary Delete user by ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.StandardResponse{data=nil} "User deleted successfully"
// @Failure 400 {object} response.StandardResponse{data=nil} "Invalid user id"
// @Failure 500 {object} response.StandardResponse{data=nil} "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.svc.RemoveUser(id); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, nil, "User deleted successfully")
}

// GetUserSettings godoc
// @Summary Get user settings
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.StandardResponse{data=user.UserSettings}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /users/{id}/settings [get]
func (h *UserHandler) GetUserSettings(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	settings, err := h.svc.GetSettings(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, settings, "User settings retrieved successfully")
}

// UpdateUserSettings godoc
// @Summary Update user settings
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param settings body map[string]interface{} true "Settings to update"
// @Success 200 {object} response.StandardResponse{data=user.UserSettings}
// @Failure 400 {object} response.StandardResponse{data=nil}
// @Failure 500 {object} response.StandardResponse{data=nil}
// @Router /users/{id}/settings [put]
func (h *UserHandler) UpdateUserSettings(c *gin.Context) {
	id, err := utils.ParseIDParam(c, "id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	// Only allow known settings fields
	allowed := map[string]bool{"theme": true, "language": true, "receive_notifications": true}
	for key := range updates {
		if !allowed[key] {
			delete(updates, key)
		}
	}

	settings, err := h.svc.UpdateSettings(c.Request.Context(), id, updates)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, settings, "User settings updated successfully")
}
