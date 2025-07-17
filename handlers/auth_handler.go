package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/middleware"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/response"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register godoc
// @Summary User registration
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param input body dto.CreateUserInput true "User registration info"
// @Success 201 {object} response.MessageResponse "User registered successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 409 {object} response.ErrorResponse "Username already taken"
// @Failure 500 {object} response.ErrorResponse "Failed to create user"
// @Router /register [post]
func Register(c *gin.Context) {
	var input dto.CreateUserInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Invalid input"})
		return
	}

	var existing models.User
	err := db.DB.Where("username = ?", input.Username).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "database error"})
		return
	}

	if err == nil {
		c.JSON(http.StatusConflict, response.ErrorResponse{Error: "Username already taken"})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := models.User{
		Username: input.Username,
		Password: string(hashed),
		Email:    input.Email,
		FullName: input.FullName,
		Type:     "origin",
		Status:   "offline",
	}

	if input.Type != nil {
		user.Type = *input.Type
	}

	if input.Status != nil {
		user.Status = *input.Status
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, response.MessageResponse{Message: "User registered successfully"})
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Success 200 {object} response.TokenResponse "JWT token and user info"
// @Failure 400 {object} response.ErrorResponse "Invalid input"
// @Failure 401 {object} response.ErrorResponse "Invalid username or password"
// @Failure 500 {object} response.ErrorResponse "Failed to generate token"
// @Router /login [post]
func Login(c *gin.Context) {
	var req struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "Invalid input"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "Invalid username or password"})
		return
	}

	token, err := middleware.GenerateToken(user.UID, user.Username, time.Hour*24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, response.TokenResponse{
		Token:    token,
		UID:      user.UID,
		Username: user.Username,
	})
}
