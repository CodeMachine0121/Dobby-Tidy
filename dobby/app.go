package main

import (
	"context"
	"log"
	"time"

	"github.com/dobby/filemanager/internal/application"
	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/dobby/filemanager/internal/query"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const scanInterval = 30 * time.Second

// App is the Wails binding thin shell.
// It receives calls from the frontend and delegates entirely to the application layer.
// No business logic lives here.
type App struct {
	ctx          context.Context
	ruleSvc      *application.RuleService
	logSvc       *application.LogService
	processorSvc *application.BackgroundProcessorService
	licenseSvc   *application.LicenseService
}

func NewApp(
	ruleSvc *application.RuleService,
	logSvc *application.LogService,
	processorSvc *application.BackgroundProcessorService,
	licenseSvc *application.LicenseService,
) *App {
	return &App{ruleSvc: ruleSvc, logSvc: logSvc, processorSvc: processorSvc, licenseSvc: licenseSvc}
}

// startup is called by Wails when the application starts.
// It initializes the trial period (if first launch) then begins the background scan loop.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.licenseSvc.InitializeTrial(ctx); err != nil {
		log.Printf("license: trial init error: %v", err)
	}
	go func() {
		// Run an initial scan immediately on startup.
		if err := a.processorSvc.ScanAndProcess(ctx); err != nil {
			log.Printf("background processor: initial scan error: %v", err)
		}
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := a.processorSvc.ScanAndProcess(ctx); err != nil {
					log.Printf("background processor: scan error: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// ── Dialogs ───────────────────────────────────────────────────────────────────

// SelectFolder opens the OS folder-picker dialog and returns the chosen path.
// Returns an empty string if the user cancels.
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "選擇資料夾",
	})
}

// ── Rules ──────────────────────────────────────────────────────────────────────

// CreateRule creates a new file-processing rule.
func (a *App) CreateRule(req CreateRuleRequest) (*RuleDTO, error) {
	r, err := a.ruleSvc.CreateRule(a.ctx, application.CreateRuleCmd{
		Name:           req.Name,
		WatchFolder:    req.WatchFolder,
		Recursive:      req.Recursive,
		FilterExts:     req.FilterExts,
		FilterKeyword:  req.FilterKeyword,
		NameTemplate:   req.NameTemplate,
		TargetTemplate: req.TargetTemplate,
		Project:        req.Project,
		TypeLabel:      req.TypeLabel,
	})
	if err != nil {
		return nil, err
	}
	return toRuleDTO(r), nil
}

// ListRules returns all rules.
func (a *App) ListRules() ([]*RuleDTO, error) {
	rules, err := a.ruleSvc.ListRules(a.ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]*RuleDTO, 0, len(rules))
	for _, r := range rules {
		dtos = append(dtos, toRuleDTO(r))
	}
	return dtos, nil
}

// GetRule returns a single rule by ID.
func (a *App) GetRule(id string) (*RuleDTO, error) {
	r, err := a.ruleSvc.GetRule(a.ctx, id)
	if err != nil {
		return nil, err
	}
	return toRuleDTO(r), nil
}

// DeleteRule removes a rule by ID.
func (a *App) DeleteRule(id string) error {
	return a.ruleSvc.DeleteRule(a.ctx, id)
}

// EnableRule activates a rule.
func (a *App) EnableRule(id string) error {
	return a.ruleSvc.EnableRule(a.ctx, id)
}

// DisableRule deactivates a rule.
func (a *App) DisableRule(id string) error {
	return a.ruleSvc.DisableRule(a.ctx, id)
}

// ── License ─────────────────────────────────────────────────────────────────────

// GetLicenseInfo returns the current license status and remaining trial days.
func (a *App) GetLicenseInfo() (*application.LicenseInfo, error) {
	return a.licenseSvc.GetLicenseInfo(a.ctx)
}

// ActivateLicense validates and activates the given license key.
func (a *App) ActivateLicense(key string) error {
	return a.licenseSvc.ActivateLicense(a.ctx, key)
}

// ── Operation Logs ─────────────────────────────────────────────────────────────

// ListRecentLogs returns the most recent operation log entries.
func (a *App) ListRecentLogs(limit int) ([]*LogDTO, error) {
	logs, err := a.logSvc.GetRecentLogs(a.ctx, limit)
	if err != nil {
		return nil, err
	}
	return toLogDTOs(logs), nil
}

// ListLogsByRule returns log entries filtered by rule ID.
func (a *App) ListLogsByRule(ruleId string, limit int) ([]*LogDTO, error) {
	logs, err := a.logSvc.GetLogsByRule(a.ctx, ruleId, limit)
	if err != nil {
		return nil, err
	}
	return toLogDTOs(logs), nil
}

// GetTodayCount returns the number of files processed today.
func (a *App) GetTodayCount() (int, error) {
	return a.logSvc.GetTodayCount(a.ctx)
}

// ── DTOs ───────────────────────────────────────────────────────────────────────

// CreateRuleRequest is the payload sent from the frontend to create a rule.
type CreateRuleRequest struct {
	Name           string   `json:"name"`
	WatchFolder    string   `json:"watchFolder"`
	Recursive      bool     `json:"recursive"`
	FilterExts     []string `json:"filterExts"`
	FilterKeyword  string   `json:"filterKeyword"`
	NameTemplate   string   `json:"nameTemplate"`
	TargetTemplate string   `json:"targetTemplate"`
	Project        string   `json:"project"`
	TypeLabel      string   `json:"typeLabel"`
}

// RuleDTO is the read representation sent to the frontend.
type RuleDTO struct {
	Id             string   `json:"id"`
	Name           string   `json:"name"`
	Enabled        bool     `json:"enabled"`
	WatchFolder    string   `json:"watchFolder"`
	Recursive      bool     `json:"recursive"`
	FilterExts     []string `json:"filterExts"`
	FilterKeyword  string   `json:"filterKeyword"`
	NameTemplate   string   `json:"nameTemplate"`
	TargetTemplate string   `json:"targetTemplate"`
	Project        string   `json:"project"`
	TypeLabel      string   `json:"typeLabel"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

// LogDTO is the read representation of an operation log entry.
type LogDTO struct {
	LogId        string `json:"logId"`
	RuleId       string `json:"ruleId"`
	RuleName     string `json:"ruleName"`
	OriginalPath string `json:"originalPath"`
	NewPath      string `json:"newPath"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	ProcessedAt  string `json:"processedAt"`
}

// ── Converters ────────────────────────────────────────────────────────────────

func toRuleDTO(r *rule.Rule) *RuleDTO {
	exts := r.FilterSpec().Extensions
	if exts == nil {
		exts = []string{}
	}
	return &RuleDTO{
		Id:             r.Id().String(),
		Name:           r.Name(),
		Enabled:        r.Enabled(),
		WatchFolder:    r.WatchConfig().FolderPath,
		Recursive:      r.WatchConfig().Recursive,
		FilterExts:     exts,
		FilterKeyword:  r.FilterSpec().Keyword,
		NameTemplate:   r.NameTemplate().TemplateString,
		TargetTemplate: r.TargetTemplate().PathTemplate,
		Project:        r.Project(),
		TypeLabel:      r.TypeLabel(),
		CreatedAt:      r.CreatedAt().UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      r.UpdatedAt().UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func toLogDTOs(logs []*query.OperationLog) []*LogDTO {
	dtos := make([]*LogDTO, 0, len(logs))
	for _, l := range logs {
		dtos = append(dtos, &LogDTO{
			LogId:        l.LogId,
			RuleId:       l.RuleId,
			RuleName:     l.RuleName,
			OriginalPath: l.OriginalPath,
			NewPath:      l.NewPath,
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
			ProcessedAt:  l.ProcessedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	return dtos
}
