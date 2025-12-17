package repositories

import (
	"github.com/linskybing/platform-go/src/db"
	"github.com/linskybing/platform-go/src/models"
)

type JobRepo interface {
	Create(job *models.Job) error
	FindAll() ([]models.Job, error)
	FindByID(id uint) (*models.Job, error)
	FindByUserID(userID uint) ([]models.Job, error)
	FindByNamespace(namespace string) ([]models.Job, error)
}

type DBJobRepo struct{}

func (r *DBJobRepo) Create(job *models.Job) error {
	return db.DB.Create(job).Error
}

func (r *DBJobRepo) FindAll() ([]models.Job, error) {
	var jobs []models.Job
	err := db.DB.Find(&jobs).Error
	return jobs, err
}

func (r *DBJobRepo) FindByID(id uint) (*models.Job, error) {
	var job models.Job
	err := db.DB.First(&job, id).Error
	return &job, err
}

func (r *DBJobRepo) FindByUserID(userID uint) ([]models.Job, error) {
	var jobs []models.Job
	err := db.DB.Where("user_id = ?", userID).Find(&jobs).Error
	return jobs, err
}

func (r *DBJobRepo) FindByNamespace(namespace string) ([]models.Job, error) {
	var jobs []models.Job
	err := db.DB.Where("namespace = ?", namespace).Find(&jobs).Error
	return jobs, err
}
