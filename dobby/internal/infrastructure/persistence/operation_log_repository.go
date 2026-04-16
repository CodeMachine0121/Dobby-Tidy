package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dobby/filemanager/internal/query"
)

// SQLiteOperationLogRepository implements both query.IOperationLogRepository
// and the read-side used by SequenceGenerator.
type SQLiteOperationLogRepository struct {
	db *sql.DB
}

func NewSQLiteOperationLogRepository(db *sql.DB) *SQLiteOperationLogRepository {
	return &SQLiteOperationLogRepository{db: db}
}

// Save persists an OperationLog entry.
func (r *SQLiteOperationLogRepository) Save(ctx context.Context, log *query.OperationLog) error {
	id := log.LogId
	if id == "" {
		id = uuid.NewString()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO operation_logs
			(id, rule_id, rule_name, original_path, new_path, status, error_message, processed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		log.RuleId,
		log.RuleName,
		log.OriginalPath,
		log.NewPath,
		log.Status,
		log.ErrorMessage,
		log.ProcessedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

// CountSuccessByRuleAndDate counts successful operations for a rule on a given calendar day (UTC).
func (r *SQLiteOperationLogRepository) CountSuccessByRuleAndDate(ctx context.Context, ruleId string, date time.Time) (int, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.UTC).Format(time.RFC3339Nano)

	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM operation_logs
		 WHERE rule_id = ?
		   AND status = 'success'
		   AND processed_at >= ?
		   AND processed_at <= ?`,
		ruleId, dayStart, dayEnd,
	).Scan(&count)
	return count, err
}

// ListRecent returns up to limit log entries ordered by processed_at DESC.
func (r *SQLiteOperationLogRepository) ListRecent(ctx context.Context, limit int) ([]*query.OperationLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, rule_id, rule_name, original_path, new_path, status, error_message, processed_at
		  FROM operation_logs
		 ORDER BY processed_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}

// ListByRule returns log entries for a specific rule ordered by processed_at DESC.
func (r *SQLiteOperationLogRepository) ListByRule(ctx context.Context, ruleId string, limit int) ([]*query.OperationLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, rule_id, rule_name, original_path, new_path, status, error_message, processed_at
		  FROM operation_logs
		 WHERE rule_id = ?
		 ORDER BY processed_at DESC
		 LIMIT ?`, ruleId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}

// CountToday counts all operations processed today (UTC).
func (r *SQLiteOperationLogRepository) CountToday(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	dayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC).Format(time.RFC3339Nano)

	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM operation_logs
		 WHERE processed_at >= ? AND processed_at <= ?`,
		dayStart, dayEnd,
	).Scan(&count)
	return count, err
}

func scanLogs(rows *sql.Rows) ([]*query.OperationLog, error) {
	var results []*query.OperationLog
	for rows.Next() {
		var (
			log          query.OperationLog
			processedStr string
		)
		if err := rows.Scan(
			&log.LogId, &log.RuleId, &log.RuleName,
			&log.OriginalPath, &log.NewPath,
			&log.Status, &log.ErrorMessage,
			&processedStr,
		); err != nil {
			return nil, fmt.Errorf("operation_log_repository scanLogs: %w", err)
		}
		log.ProcessedAt, _ = time.Parse(time.RFC3339Nano, processedStr)
		results = append(results, &log)
	}
	return results, rows.Err()
}
