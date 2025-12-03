package services

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/linskybing/platform-go/src/models"
	"github.com/linskybing/platform-go/src/repositories"
	"github.com/linskybing/platform-go/src/repositories/mock_repositories"
)

func TestAuditService_QueryAuditLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAudit := mock_repositories.NewMockAuditRepo(ctrl)
	repos := &repositories.Repos{Audit: mockAudit}
	svc := NewAuditService(repos)

	params := repositories.AuditQueryParams{Limit: 10}

	expected := []models.AuditLog{{ID: 1}}
	mockAudit.EXPECT().GetAuditLogs(params).Return(expected, nil)

	result, err := svc.QueryAuditLogs(params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 || result[0].ID != 1 {
		t.Fatalf("expected one audit log with ID=1, got %+v", result)
	}

	mockAudit.EXPECT().GetAuditLogs(params).Return(nil, errors.New("db error"))
	_, err = svc.QueryAuditLogs(params)
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", err)
	}
}
