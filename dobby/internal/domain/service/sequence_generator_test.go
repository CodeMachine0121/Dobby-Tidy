package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/dobby/filemanager/internal/domain/service"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-14: 當日尚無任何成功紀錄時，序號從 001 開始
// ─────────────────────────────────────────────────────────────────────────────

// S-14-1
func TestSequenceGenerator_NoHistory_ReturnsFirstSeq(t *testing.T) {
	// Arrange
	logRepo := givenNoSuccessLogs()
	gen := service.NewSequenceGenerator(logRepo)

	// Act
	seq, err := gen.Generate(context.Background(), "rule-001", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "001", seq)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-15: 當日已有 3 筆成功紀錄時，序號為 004
// ─────────────────────────────────────────────────────────────────────────────

// S-15-1
func TestSequenceGenerator_ThreeExisting_ReturnsFourthSeq(t *testing.T) {
	// Arrange
	logRepo := givenSuccessLogCount(3)
	gen := service.NewSequenceGenerator(logRepo)

	// Act
	seq, err := gen.Generate(context.Background(), "rule-001", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC))

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "004", seq)
}

// ─────────────────────────────────────────────────────────────────────────────
// Mock: IOperationLogRepository
// ─────────────────────────────────────────────────────────────────────────────

type mockOperationLogRepo struct {
	count int
}

func (m *mockOperationLogRepo) CountSuccessByRuleAndDate(_ context.Context, _ string, _ time.Time) (int, error) {
	return m.count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func givenNoSuccessLogs() service.IOperationLogRepository {
	return &mockOperationLogRepo{count: 0}
}

func givenSuccessLogCount(n int) service.IOperationLogRepository {
	return &mockOperationLogRepo{count: n}
}
