package query

import (
	"context"
	"time"
)

// OperationLog is the read model for a processed file event.
// Written by the infrastructure layer; read by the query layer and API layer.
type OperationLog struct {
	LogId        string
	RuleId       string
	RuleName     string
	OriginalPath string
	NewPath      string
	Status       string // "success" | "error"
	ErrorMessage string
	ProcessedAt  time.Time
}

// IOperationLogRepository is the full contract for the operation_logs table,
// covering both write operations and all read queries.
type IOperationLogRepository interface {
	Save(ctx context.Context, log *OperationLog) error
	CountSuccessByRuleAndDate(ctx context.Context, ruleId string, date time.Time) (int, error)
	ListRecent(ctx context.Context, limit int) ([]*OperationLog, error)
	ListByRule(ctx context.Context, ruleId string, limit int) ([]*OperationLog, error)
	CountToday(ctx context.Context) (int, error)
}
