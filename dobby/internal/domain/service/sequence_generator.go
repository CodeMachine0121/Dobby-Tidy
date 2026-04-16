package service

import (
	"context"
	"fmt"
	"time"
)

// IOperationLogRepository is the read-side contract needed by SequenceGenerator.
type IOperationLogRepository interface {
	CountSuccessByRuleAndDate(ctx context.Context, ruleId string, date time.Time) (int, error)
}

// SequenceGenerator is a Domain Service that produces zero-padded 3-digit sequence numbers
// based on how many files a given rule has successfully processed today.
type SequenceGenerator struct {
	logRepo IOperationLogRepository
}

func NewSequenceGenerator(logRepo IOperationLogRepository) *SequenceGenerator {
	return &SequenceGenerator{logRepo: logRepo}
}

// Generate returns a sequence string like "001", "002", ... for the given rule and date.
func (g *SequenceGenerator) Generate(ctx context.Context, ruleId string, date time.Time) (string, error) {
	count, err := g.logRepo.CountSuccessByRuleAndDate(ctx, ruleId, date)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%03d", count+1), nil
}
