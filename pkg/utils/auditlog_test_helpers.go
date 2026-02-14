package utils

import (
	"sync"

	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type mockAuditRepo struct {
	mu          sync.Mutex
	createdLogs []*audit.AuditLog
	shouldErr   bool
}

func (m *mockAuditRepo) CreateAuditLog(log *audit.AuditLog) error {
	if m.shouldErr {
		return nil // Simulating error (we ignore it in LogAudit)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createdLogs == nil {
		m.createdLogs = make([]*audit.AuditLog, 0)
	}
	m.createdLogs = append(m.createdLogs, log)
	return nil
}

func (m *mockAuditRepo) GetAuditLogs(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepo) DeleteOldAuditLogs(retentionDays int) error {
	return nil
}

func (m *mockAuditRepo) WithTx(tx *gorm.DB) repository.AuditRepo {
	return m
}
