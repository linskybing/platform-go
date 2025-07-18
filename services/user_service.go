package services

import (
	"errors"
	"time"

	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/middleware"
	"github.com/linskybing/platform-go/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RegisterUser(input dto.CreateUserInput) error {
	var existing models.User
	err := db.DB.Where("username = ?", input.Username).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		return errors.New("username already taken")
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

	return db.DB.Create(&user).Error
}

func LoginUser(username, password string) (models.User, string, error) {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return user, "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return user, "", errors.New("invalid credentials")
	}

	token, err := middleware.GenerateToken(user.UID, user.Username, 24*time.Hour)
	if err != nil {
		return user, "", err
	}

	return user, token, nil
}
