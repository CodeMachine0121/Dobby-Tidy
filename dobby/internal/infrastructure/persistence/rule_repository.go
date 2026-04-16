package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dobby/filemanager/internal/domain/rule"
)

// SQLiteRuleRepository implements rule.IRuleRepository using SQLite.
type SQLiteRuleRepository struct {
	db *sql.DB
}

func NewSQLiteRuleRepository(db *sql.DB) *SQLiteRuleRepository {
	return &SQLiteRuleRepository{db: db}
}

// Save upserts a Rule (INSERT OR REPLACE).
func (r *SQLiteRuleRepository) Save(ctx context.Context, rl *rule.Rule) error {
	exts, err := json.Marshal(rl.FilterSpec().Extensions)
	if err != nil {
		return fmt.Errorf("rule_repository.Save marshal extensions: %w", err)
	}

	enabled := 0
	if rl.Enabled() {
		enabled = 1
	}
	recursive := 0
	if rl.WatchConfig().Recursive {
		recursive = 1
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO rules
			(id, name, enabled, watch_folder, recursive,
			 filter_extensions, filter_keyword,
			 name_template, target_template,
			 project, type_label, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name             = excluded.name,
			enabled          = excluded.enabled,
			watch_folder     = excluded.watch_folder,
			recursive        = excluded.recursive,
			filter_extensions= excluded.filter_extensions,
			filter_keyword   = excluded.filter_keyword,
			name_template    = excluded.name_template,
			target_template  = excluded.target_template,
			project          = excluded.project,
			type_label       = excluded.type_label,
			updated_at       = excluded.updated_at`,
		rl.Id().String(),
		rl.Name(),
		enabled,
		rl.WatchConfig().FolderPath,
		recursive,
		string(exts),
		rl.FilterSpec().Keyword,
		rl.NameTemplate().TemplateString,
		rl.TargetTemplate().PathTemplate,
		rl.Project(),
		rl.TypeLabel(),
		rl.CreatedAt().UTC().Format(time.RFC3339Nano),
		rl.UpdatedAt().UTC().Format(time.RFC3339Nano),
	)
	return err
}

// FindById retrieves a Rule by its ID. Returns nil, nil when not found.
func (r *SQLiteRuleRepository) FindById(ctx context.Context, id rule.RuleId) (*rule.Rule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, enabled, watch_folder, recursive,
		        filter_extensions, filter_keyword,
		        name_template, target_template,
		        project, type_label, created_at, updated_at
		   FROM rules WHERE id = ?`, id.String())
	return scanRule(row)
}

// FindByFolderPath retrieves a Rule by its watch_folder. Returns nil, nil when not found.
func (r *SQLiteRuleRepository) FindByFolderPath(ctx context.Context, folderPath string) (*rule.Rule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, enabled, watch_folder, recursive,
		        filter_extensions, filter_keyword,
		        name_template, target_template,
		        project, type_label, created_at, updated_at
		   FROM rules WHERE watch_folder = ?`, folderPath)
	return scanRule(row)
}

// ExistsByFolderPath returns true when a rule with the given watch_folder exists.
func (r *SQLiteRuleRepository) ExistsByFolderPath(ctx context.Context, folderPath string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rules WHERE watch_folder = ?`, folderPath,
	).Scan(&count)
	return count > 0, err
}

// Delete removes a Rule by ID.
func (r *SQLiteRuleRepository) Delete(ctx context.Context, id rule.RuleId) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM rules WHERE id = ?`, id.String())
	return err
}

// ListAll returns all rules ordered by created_at ASC.
func (r *SQLiteRuleRepository) ListAll(ctx context.Context) ([]*rule.Rule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, enabled, watch_folder, recursive,
		       filter_extensions, filter_keyword,
		       name_template, target_template,
		       project, type_label, created_at, updated_at
		  FROM rules
		 ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*rule.Rule
	for rows.Next() {
		rl, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		if rl != nil {
			results = append(results, rl)
		}
	}
	return results, rows.Err()
}

// ─── scanner ─────────────────────────────────────────────────────────────────

type ruleScanner interface {
	Scan(dest ...any) error
}

func scanRule(scanner ruleScanner) (*rule.Rule, error) {
	var (
		idStr, name, watchFolder     string
		enabledInt, recursiveInt     int
		filterExtJSON, filterKeyword string
		nameTemplate, targetTemplate string
		project, typeLabel           string
		createdAtStr, updatedAtStr   string
	)

	err := scanner.Scan(
		&idStr, &name, &enabledInt, &watchFolder, &recursiveInt,
		&filterExtJSON, &filterKeyword,
		&nameTemplate, &targetTemplate,
		&project, &typeLabel,
		&createdAtStr, &updatedAtStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("rule_repository scanRule: %w", err)
	}

	var extensions []string
	if err := json.Unmarshal([]byte(filterExtJSON), &extensions); err != nil {
		extensions = nil
	}

	createdAt, _ := time.Parse(time.RFC3339Nano, createdAtStr)
	updatedAt, _ := time.Parse(time.RFC3339Nano, updatedAtStr)

	return rule.Reconstitute(
		rule.RuleIdFrom(idStr),
		name,
		enabledInt == 1,
		rule.WatchConfig{FolderPath: watchFolder, Recursive: recursiveInt == 1},
		rule.FilterSpec{Extensions: extensions, Keyword: filterKeyword},
		rule.NamingTemplate{TemplateString: nameTemplate},
		rule.TargetPathTemplate{PathTemplate: targetTemplate},
		project,
		typeLabel,
		createdAt,
		updatedAt,
	), nil
}
