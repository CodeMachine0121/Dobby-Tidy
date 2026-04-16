package application

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"time"

	domainservice "github.com/dobby/filemanager/internal/domain/service"

	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/dobby/filemanager/internal/query"
)

// ─────────────────────────────────────────────────────────────────────────────
// Ports (defined here so infrastructure depends on application, not the reverse)
// ─────────────────────────────────────────────────────────────────────────────

// IFileSystem abstracts file-system operations so they can be mocked in tests.
type IFileSystem interface {
	// ListFiles returns all files (not directories) under dir.
	// When recursive is true, subdirectories are traversed as well.
	ListFiles(ctx context.Context, dir string, recursive bool) ([]FileInfo, error)

	// MoveFile moves (and renames) the file at src to dst.
	MoveFile(ctx context.Context, src, dst string) error

	// EnsureDir creates dir and all necessary parents (equivalent to os.MkdirAll).
	EnsureDir(ctx context.Context, dir string) error
}

// FileInfo describes a single file discovered by IFileSystem.ListFiles.
type FileInfo struct {
	Path       string // full absolute path
	Name       string // base name without extension
	Extension  string // includes the dot, e.g. ".pdf"
	DetectedAt time.Time
}

// IOperationLogWriter is the write-side contract for operation logs.
type IOperationLogWriter interface {
	Save(ctx context.Context, log *query.OperationLog) error
}

// ─────────────────────────────────────────────────────────────────────────────
// BackgroundProcessorService
// ─────────────────────────────────────────────────────────────────────────────

// BackgroundProcessorService scans watched folders and processes files that
// match enabled rules. Callers are responsible for scheduling (e.g. a ticker).
type BackgroundProcessorService struct {
	ruleRepo  rule.IRuleRepository
	jobRepo   job.IProcessingJobRepository
	logWriter IOperationLogWriter
	fs        IFileSystem
	matcher   *domainservice.RuleMatcher
	renderer  *domainservice.TemplateRenderer
	seqGen    *domainservice.SequenceGenerator
	running   atomic.Bool
}

// NewBackgroundProcessorService wires up the service with its dependencies.
func NewBackgroundProcessorService(
	ruleRepo rule.IRuleRepository,
	jobRepo job.IProcessingJobRepository,
	logWriter IOperationLogWriter,
	fs IFileSystem,
	seqLogRepo domainservice.IOperationLogRepository,
) *BackgroundProcessorService {
	return &BackgroundProcessorService{
		ruleRepo:  ruleRepo,
		jobRepo:   jobRepo,
		logWriter: logWriter,
		fs:        fs,
		matcher:   domainservice.NewRuleMatcher(),
		renderer:  domainservice.NewTemplateRenderer(),
		seqGen:    domainservice.NewSequenceGenerator(seqLogRepo),
	}
}

// ForceRunning marks the service as already running. Used in tests only.
func (s *BackgroundProcessorService) ForceRunning() {
	s.running.Store(true)
}

// ScanAndProcess scans all enabled rules' watch folders and processes any
// files that satisfy their filter specs. If a scan is already in progress the
// call returns immediately without error (S-6).
func (s *BackgroundProcessorService) ScanAndProcess(ctx context.Context) error {
	if !s.running.CompareAndSwap(false, true) {
		return nil
	}
	defer s.running.Store(false)

	rules, err := s.ruleRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, r := range rules {
		if !r.Enabled() {
			continue
		}
		if err := s.processRule(ctx, r); err != nil {
			// Log the error but continue processing other rules.
			_ = err
		}
	}
	return nil
}

func (s *BackgroundProcessorService) processRule(ctx context.Context, r *rule.Rule) error {
	cfg := r.WatchConfig()
	files, err := s.fs.ListFiles(ctx, cfg.FolderPath, cfg.Recursive)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !s.matcher.Match(f.Extension, f.Name, r.FilterSpec()) {
			continue
		}
		s.processFile(ctx, r, f)
	}
	return nil
}

func (s *BackgroundProcessorService) processFile(ctx context.Context, r *rule.Rule, f FileInfo) {
	now := time.Now()

	fileEvent := job.FileEvent{
		DetectedPath: f.Path,
		OriginalName: f.Name,
		Extension:    f.Extension,
		DetectedAt:   f.DetectedAt,
	}

	j := job.NewProcessingJob(fileEvent)
	_ = s.jobRepo.Save(ctx, j)

	seq, err := s.seqGen.Generate(ctx, r.Id().String(), now)
	if err != nil {
		seq = "001"
	}

	procCtx := job.ProcessingContext{
		Project:      r.Project(),
		TypeLabel:    r.TypeLabel(),
		Date:         now,
		Seq:          seq,
		OriginalName: f.Name,
		Extension:    f.Extension,
	}

	_ = j.MarkMatched(r.Id(), procCtx)
	_ = j.MarkProcessing()
	_ = s.jobRepo.Save(ctx, j)

	newName := s.renderer.RenderName(r.NameTemplate(), procCtx)
	targetDir := s.renderer.RenderPath(r.TargetTemplate(), procCtx)
	targetPath := filepath.Join(targetDir, newName)

	_ = s.fs.EnsureDir(ctx, targetDir)
	moveErr := s.fs.MoveFile(ctx, f.Path, targetPath)

	var logStatus, logErrMsg, logNewPath string
	if moveErr != nil {
		_ = j.MarkFailed(moveErr.Error())
		logStatus = "failed"
		logErrMsg = moveErr.Error()
	} else {
		_ = j.MarkSucceeded(targetPath)
		logStatus = "success"
		logNewPath = targetPath
	}

	_ = s.jobRepo.Save(ctx, j)

	_ = s.logWriter.Save(ctx, &query.OperationLog{
		RuleId:       r.Id().String(),
		RuleName:     r.Name(),
		OriginalPath: f.Path,
		NewPath:      logNewPath,
		Status:       logStatus,
		ErrorMessage: logErrMsg,
		ProcessedAt:  now,
	})
}
