package services

import (
	"github.com/linskybing/platform-go/src/models"
	"github.com/linskybing/platform-go/src/repositories"
)

type AuditService struct {
	Repos *repositories.Repos
}

func NewAuditService(repos *repositories.Repos) *AuditService {
	return &AuditService{
		Repos: repos,
	}
}

func (s *AuditService) QueryAuditLogs(params repositories.AuditQueryParams) ([]models.AuditLog, error) {
	return s.Repos.Audit.GetAuditLogs(params)
}
