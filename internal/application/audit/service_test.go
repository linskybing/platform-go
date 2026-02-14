package audit

import (
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubAuditRepo struct {
	getAuditLogs func(params repository.AuditQueryParams) ([]audit.AuditLog, error)
}

func (s *stubAuditRepo) GetAuditLogs(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
	if s.getAuditLogs != nil {
		return s.getAuditLogs(params)
	}
	return nil, nil
}

func (s *stubAuditRepo) CreateAuditLog(a *audit.AuditLog) error     { return nil }
func (s *stubAuditRepo) DeleteOldAuditLogs(retentionDays int) error { return nil }
func (s *stubAuditRepo) WithTx(tx *gorm.DB) repository.AuditRepo    { return s }

func TestAuditService_QueryAuditLogs(t *testing.T) {
	stubAudit := &stubAuditRepo{}
	repos := &repository.Repos{Audit: stubAudit}
	svc := NewAuditService(repos)

	params := repository.AuditQueryParams{Limit: 10}

	expected := []audit.AuditLog{{ID: 1}}
	stubAudit.getAuditLogs = func(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
		return expected, nil
	}

	result, err := svc.QueryAuditLogs(params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 || result[0].ID != 1 {
		t.Fatalf("expected one audit log with ID=1, got %+v", result)
	}

	stubAudit.getAuditLogs = func(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
		return nil, errors.New("db error")
	}
	_, err = svc.QueryAuditLogs(params)
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", err)
	}
}
