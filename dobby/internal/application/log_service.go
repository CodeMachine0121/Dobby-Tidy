package application

import (
	"context"
	"fmt"

	"github.com/dobby/filemanager/internal/query"
)

// LogService handles read operations on operation logs.
type LogService struct {
	logRepo query.IOperationLogRepository
}

func NewLogService(logRepo query.IOperationLogRepository) *LogService {
	return &LogService{logRepo: logRepo}
}

// GetRecentLogs returns the most recent operation log entries (up to limit).
func (s *LogService) GetRecentLogs(ctx context.Context, limit int) ([]*query.OperationLog, error) {
	if limit <= 0 {
		limit = 50
	}
	logs, err := s.logRepo.ListRecent(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("log_service.GetRecentLogs: %w", err)
	}
	return logs, nil
}

// GetLogsByRule returns log entries for a specific rule.
func (s *LogService) GetLogsByRule(ctx context.Context, ruleId string, limit int) ([]*query.OperationLog, error) {
	if limit <= 0 {
		limit = 50
	}
	logs, err := s.logRepo.ListByRule(ctx, ruleId, limit)
	if err != nil {
		return nil, fmt.Errorf("log_service.GetLogsByRule: %w", err)
	}
	return logs, nil
}

// GetTodayCount returns the total number of operations processed today.
func (s *LogService) GetTodayCount(ctx context.Context) (int, error) {
	count, err := s.logRepo.CountToday(ctx)
	if err != nil {
		return 0, fmt.Errorf("log_service.GetTodayCount: %w", err)
	}
	return count, nil
}
