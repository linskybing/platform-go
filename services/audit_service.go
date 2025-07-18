package services

import (
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
)

func QueryAuditLogs(params repositories.AuditQueryParams) ([]models.AuditLog, error) {
	return repositories.GetAuditLogs(params)
}
