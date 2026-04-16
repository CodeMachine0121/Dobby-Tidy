package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
)

// SQLiteProcessingJobRepository implements job.IProcessingJobRepository using SQLite.
type SQLiteProcessingJobRepository struct {
	db *sql.DB
}

func NewSQLiteProcessingJobRepository(db *sql.DB) *SQLiteProcessingJobRepository {
	return &SQLiteProcessingJobRepository{db: db}
}

// Save upserts a ProcessingJob.
func (r *SQLiteProcessingJobRepository) Save(ctx context.Context, j *job.ProcessingJob) error {
	var (
		ctxProject, ctxTypeLabel, ctxSeq, ctxOrigName, ctxExt sql.NullString
		ctxDate                                               sql.NullString
		resultNewPath, resultErrMsg, resultProcessedAt        sql.NullString
	)

	if c := j.Context(); c != nil {
		ctxProject = nullStr(c.Project)
		ctxTypeLabel = nullStr(c.TypeLabel)
		ctxSeq = nullStr(c.Seq)
		ctxOrigName = nullStr(c.OriginalName)
		ctxExt = nullStr(c.Extension)
		ctxDate = nullStr(c.Date.UTC().Format(time.RFC3339Nano))
	}
	if res := j.Result(); res != nil {
		resultNewPath = nullStr(res.NewPath)
		resultErrMsg = nullStr(res.ErrorMessage)
		resultProcessedAt = nullStr(res.ProcessedAt.UTC().Format(time.RFC3339Nano))
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO processing_jobs
			(id, rule_id,
			 file_event_path, file_event_name, file_event_extension, file_event_detected_at,
			 state,
			 ctx_project, ctx_type_label, ctx_date, ctx_seq, ctx_original_name, ctx_extension,
			 result_new_path, result_error_message, result_processed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			rule_id                = excluded.rule_id,
			state                  = excluded.state,
			ctx_project            = excluded.ctx_project,
			ctx_type_label         = excluded.ctx_type_label,
			ctx_date               = excluded.ctx_date,
			ctx_seq                = excluded.ctx_seq,
			ctx_original_name      = excluded.ctx_original_name,
			ctx_extension          = excluded.ctx_extension,
			result_new_path        = excluded.result_new_path,
			result_error_message   = excluded.result_error_message,
			result_processed_at    = excluded.result_processed_at`,
		j.Id().String(),
		j.RuleId().String(),
		j.FileEvent().DetectedPath,
		j.FileEvent().OriginalName,
		j.FileEvent().Extension,
		j.FileEvent().DetectedAt.UTC().Format(time.RFC3339Nano),
		string(j.State()),
		ctxProject, ctxTypeLabel, ctxDate, ctxSeq, ctxOrigName, ctxExt,
		resultNewPath, resultErrMsg, resultProcessedAt,
	)
	return err
}

// FindById retrieves a ProcessingJob by ID. Returns nil, nil when not found.
func (r *SQLiteProcessingJobRepository) FindById(ctx context.Context, id job.JobId) (*job.ProcessingJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, rule_id,
		       file_event_path, file_event_name, file_event_extension, file_event_detected_at,
		       state,
		       ctx_project, ctx_type_label, ctx_date, ctx_seq, ctx_original_name, ctx_extension,
		       result_new_path, result_error_message, result_processed_at
		  FROM processing_jobs WHERE id = ?`, id.String())
	return scanJob(row)
}

// ─── scanner ─────────────────────────────────────────────────────────────────

type jobScanner interface {
	Scan(dest ...any) error
}

func scanJob(scanner jobScanner) (*job.ProcessingJob, error) {
	var (
		idStr, ruleIdStr                                         string
		evPath, evName, evExt, evDetectedAtStr                   string
		stateStr                                                 string
		ctxProject, ctxTypeLabel, ctxSeq, ctxOrigName, ctxExtStr sql.NullString
		ctxDateStr                                               sql.NullString
		resultNewPath, resultErrMsg, resultProcessedAt           sql.NullString
	)

	err := scanner.Scan(
		&idStr, &ruleIdStr,
		&evPath, &evName, &evExt, &evDetectedAtStr,
		&stateStr,
		&ctxProject, &ctxTypeLabel, &ctxDateStr, &ctxSeq, &ctxOrigName, &ctxExtStr,
		&resultNewPath, &resultErrMsg, &resultProcessedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("job_repository scanJob: %w", err)
	}

	detectedAt, _ := time.Parse(time.RFC3339Nano, evDetectedAtStr)
	fileEvent := job.FileEvent{
		DetectedPath: evPath,
		OriginalName: evName,
		Extension:    evExt,
		DetectedAt:   detectedAt,
	}

	var processingCtx *job.ProcessingContext
	if ctxProject.Valid {
		ctxDate, _ := time.Parse(time.RFC3339Nano, ctxDateStr.String)
		processingCtx = &job.ProcessingContext{
			Project:      ctxProject.String,
			TypeLabel:    ctxTypeLabel.String,
			Date:         ctxDate,
			Seq:          ctxSeq.String,
			OriginalName: ctxOrigName.String,
			Extension:    ctxExtStr.String,
		}
	}

	var processingResult *job.ProcessingResult
	if resultProcessedAt.Valid {
		processedAt, _ := time.Parse(time.RFC3339Nano, resultProcessedAt.String)
		processingResult = &job.ProcessingResult{
			NewPath:      resultNewPath.String,
			ErrorMessage: resultErrMsg.String,
			ProcessedAt:  processedAt,
		}
	}

	return job.Reconstitute(
		job.JobIdFrom(idStr),
		rule.RuleIdFrom(ruleIdStr),
		fileEvent,
		job.JobState(stateStr),
		processingCtx,
		processingResult,
	), nil
}

func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
