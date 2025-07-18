package repositories

import (
	"github.com/linskybing/platform-go/db"
	"github.com/linskybing/platform-go/models"
)

func GetProjectByID(id string) (models.Project, error) {
	var project models.Project
	err := db.DB.First(&project, id).Error
	return project, err
}

func CreateProject(p *models.Project) error {
	return db.DB.Create(p).Error
}

func UpdateProject(p *models.Project) error {
	return db.DB.Save(p).Error
}

func DeleteProject(id string) error {
	return db.DB.Delete(&models.Project{}, id).Error
}

func ListProjects() ([]models.Project, error) {
	var projects []models.Project
	err := db.DB.Find(&projects).Error
	return projects, err
}
