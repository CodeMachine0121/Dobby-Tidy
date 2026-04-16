package persistence

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// migrations is the ordered list of DDL statements applied once on startup (code-first).
// Each entry is identified by its 1-based index. Never edit existing entries; only append.
var migrations = []string{
	// v1 — initial schema
	`CREATE TABLE IF NOT EXISTS rules (
		id               TEXT    PRIMARY KEY,
		name             TEXT    NOT NULL,
		enabled          INTEGER NOT NULL DEFAULT 1,
		watch_folder     TEXT    NOT NULL UNIQUE,
		recursive        INTEGER NOT NULL DEFAULT 0,
		filter_extensions TEXT   NOT NULL DEFAULT '[]',
		filter_keyword   TEXT    NOT NULL DEFAULT '',
		name_template    TEXT    NOT NULL,
		target_template  TEXT    NOT NULL,
		project          TEXT    NOT NULL,
		type_label       TEXT    NOT NULL,
		created_at       TEXT    NOT NULL,
		updated_at       TEXT    NOT NULL
	)`,

	`CREATE TABLE IF NOT EXISTS processing_jobs (
		id                       TEXT PRIMARY KEY,
		rule_id                  TEXT NOT NULL,
		file_event_path          TEXT NOT NULL,
		file_event_name          TEXT NOT NULL,
		file_event_extension     TEXT NOT NULL,
		file_event_detected_at   TEXT NOT NULL,
		state                    TEXT NOT NULL,
		ctx_project              TEXT,
		ctx_type_label           TEXT,
		ctx_date                 TEXT,
		ctx_seq                  TEXT,
		ctx_original_name        TEXT,
		ctx_extension            TEXT,
		result_new_path          TEXT,
		result_error_message     TEXT,
		result_processed_at      TEXT
	)`,

	`CREATE TABLE IF NOT EXISTS operation_logs (
		id            TEXT PRIMARY KEY,
		rule_id       TEXT NOT NULL,
		rule_name     TEXT NOT NULL,
		original_path TEXT NOT NULL,
		new_path      TEXT NOT NULL DEFAULT '',
		status        TEXT NOT NULL,
		error_message TEXT NOT NULL DEFAULT '',
		processed_at  TEXT NOT NULL
	)`,

	`CREATE INDEX IF NOT EXISTS idx_operation_logs_rule_date
		ON operation_logs (rule_id, processed_at)`,

	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER PRIMARY KEY,
		applied_at TEXT    NOT NULL DEFAULT (datetime('now'))
	)`,
}

// Open opens (or creates) the SQLite database at dsn and applies pending migrations.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("persistence.Open: %w", err)
	}
	// SQLite works best single-writer; keep connection pool minimal.
	db.SetMaxOpenConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("persistence.Open migrations: %w", err)
	}
	return db, nil
}

// runMigrations applies any migration whose version has not yet been recorded.
func runMigrations(db *sql.DB) error {
	ctx := context.Background()

	// Bootstrap the migrations tracking table before anything else.
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)`,
	); err != nil {
		return err
	}

	for i, ddl := range migrations {
		version := i + 1

		var exists int
		err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("migration v%d check: %w", version, err)
		}
		if exists > 0 {
			continue
		}

		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("migration v%d apply: %w", version, err)
		}
		if _, err := db.ExecContext(ctx,
			`INSERT INTO schema_migrations (version) VALUES (?)`, version,
		); err != nil {
			return fmt.Errorf("migration v%d record: %w", version, err)
		}
	}
	return nil
}
