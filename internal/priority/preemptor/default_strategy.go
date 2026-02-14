package preemptor

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// SQLStrategy implements preemption using PostgreSQL window functions.
type SQLStrategy struct {
	db *gorm.DB
}

// NewSQLStrategy creates a new SQL-based preemption strategy.
func NewSQLStrategy(db *gorm.DB) *SQLStrategy {
	return &SQLStrategy{db: db}
}

// Name returns the strategy name.
func (s *SQLStrategy) Name() string {
	return "sql-preemption"
}

// Execute finds victim jobs using cumulative resource analysis.
func (s *SQLStrategy) Execute(ctx context.Context, req ResourceRequirement) (*PreemptionDecision, error) {
	var victims []struct {
		ID uint `gorm:"column:id"`
	}

	query := `
		WITH running_victims AS (
			SELECT j.id, pc.value as priority_value, j.required_gpu,
				SUM(j.required_gpu) OVER (ORDER BY pc.value ASC, j.started_at DESC) as cumulative_gpu
			FROM jobs j
			JOIN priority_classes pc ON j.priority_class_name = pc.name
			WHERE j.status = 'RUNNING' AND pc.value < ?
		)
		SELECT id FROM running_victims WHERE cumulative_gpu <= ? + (SELECT required_gpu FROM running_victims LIMIT 1)
	`
	// Note: The logic above is a simplified version of the one in the proposal.
	// In a real scenario, we might want to be more precise about the GPU target.

	err := s.db.WithContext(ctx).Raw(query, req.PriorityValue, req.GPU).Scan(&victims).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find victims: %w", err)
	}

	decision := &PreemptionDecision{
		Reason: "High priority task resource requirement",
	}
	for _, v := range victims {
		decision.JobsToPreempt = append(decision.JobsToPreempt, v.ID)
	}

	return decision, nil
}
