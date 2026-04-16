package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dobby/filemanager/internal/application"
	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/dobby/filemanager/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-1: Scan discovers a matching file and processes it successfully
// ─────────────────────────────────────────────────────────────────────────────

// S-1-1: ProcessingJob 最終狀態為 succeeded
func TestBackgroundProcessor_MatchingFile_JobStateIsSucceeded(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)

	// Act
	err := svc.ScanAndProcess(context.Background())

	// Assert
	assert.NoError(t, err)
}

// S-1-2: MoveFile 被呼叫
func TestBackgroundProcessor_MatchingFile_MoveFileIsCalled(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	fs.AssertCalled(t, "MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"))
}

// S-1-3: OperationLog 以 "success" 狀態被寫入
func TestBackgroundProcessor_MatchingFile_LogWrittenWithStatusSuccess(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)

	var capturedLog *query.OperationLog
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).
		Run(func(args mock.Arguments) { capturedLog = args.Get(1).(*query.OperationLog) }).
		Return(nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	assert.Equal(t, "success", capturedLog.Status)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-2: Scan skips files that do not match the extension filter
// ─────────────────────────────────────────────────────────────────────────────

// S-2-1: 副檔名不符時，JobRepository.Save 不被呼叫
func TestBackgroundProcessor_ExtensionMismatch_JobNotSaved(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, jobRepo, _, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createJpgFileInfo()})

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	jobRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// S-2-2: 副檔名不符時，OperationLogWriter.Save 不被呼叫
func TestBackgroundProcessor_ExtensionMismatch_LogNotWritten(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, _, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createJpgFileInfo()})

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	logWriter.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-3: Scan skips files that do not match the keyword filter
// ─────────────────────────────────────────────────────────────────────────────

// S-3-1: keyword 不符時，JobRepository.Save 不被呼叫
func TestBackgroundProcessor_KeywordMismatch_JobNotSaved(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, jobRepo, _, fs := createBackgroundProcessorService()
	givenEnabledRuleWithKeyword(ruleRepo, "contract")
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()}) // name: invoice.pdf

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	jobRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-4: File move failure marks job as failed and writes error log
// ─────────────────────────────────────────────────────────────────────────────

// S-4-1: 移動失敗時，OperationLog 以 "failed" 狀態被寫入
func TestBackgroundProcessor_MoveFails_LogWrittenWithStatusFailed(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(errors.New("permission denied"))
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)

	var capturedLog *query.OperationLog
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).
		Run(func(args mock.Arguments) { capturedLog = args.Get(1).(*query.OperationLog) }).
		Return(nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	assert.Equal(t, "failed", capturedLog.Status)
}

// S-4-2: 移動失敗時，錯誤訊息包含原始錯誤
func TestBackgroundProcessor_MoveFails_ErrorMessageContainsOriginalError(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(errors.New("permission denied"))
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)

	var capturedLog *query.OperationLog
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).
		Run(func(args mock.Arguments) { capturedLog = args.Get(1).(*query.OperationLog) }).
		Return(nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	assert.Contains(t, capturedLog.ErrorMessage, "permission denied")
}

// ─────────────────────────────────────────────────────────────────────────────
// S-5: No enabled rules results in no processing
// ─────────────────────────────────────────────────────────────────────────────

// S-5-1: 無啟用規則時，JobRepository.Save 不被呼叫
func TestBackgroundProcessor_NoEnabledRules_JobNotSaved(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, jobRepo, _, _ := createBackgroundProcessorService()
	givenNoEnabledRules(ruleRepo)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	jobRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// S-5-2: 無啟用規則時，回傳 nil 錯誤
func TestBackgroundProcessor_NoEnabledRules_ReturnsNoError(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, _, _, _ := createBackgroundProcessorService()
	givenNoEnabledRules(ruleRepo)

	// Act
	err := svc.ScanAndProcess(context.Background())

	// Assert
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-6: Scan is skipped when another scan is already running
// ─────────────────────────────────────────────────────────────────────────────

// S-6-1: 掃描進行中再次呼叫，不回傳錯誤
func TestBackgroundProcessor_AlreadyRunning_ReturnsNoError(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)

	svc.ForceRunning() // 模擬正在掃描

	// Act
	err := svc.ScanAndProcess(context.Background())

	// Assert
	assert.NoError(t, err)
}

// S-6-2: 掃描進行中再次呼叫，不掃描檔案（ListFiles 不被呼叫）
func TestBackgroundProcessor_AlreadyRunning_DoesNotListFiles(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, _, _, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)

	svc.ForceRunning() // 模擬正在掃描

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	fs.AssertNotCalled(t, "ListFiles", mock.Anything, mock.Anything, mock.Anything)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-7: Scan covers recursive subdirectories when rule is recursive
// ─────────────────────────────────────────────────────────────────────────────

// S-7-1: recursive=true 時，ListFiles 以 recursive=true 呼叫
func TestBackgroundProcessor_RecursiveRule_ListFilesCalledWithRecursiveTrue(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRecursiveRule(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", true, []application.FileInfo{createPdfSubdirFileInfo()})
	givenSeqIsFirst(seqLogRepo)
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	fs.AssertCalled(t, "ListFiles", mock.Anything, "/Downloads", true)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-8: Scan ignores files in subdirectories when rule is non-recursive
// ─────────────────────────────────────────────────────────────────────────────

// S-8-1: recursive=false 時，ListFiles 以 recursive=false 呼叫
func TestBackgroundProcessor_NonRecursiveRule_ListFilesCalledWithRecursiveFalse(t *testing.T) {
	// Arrange
	svc, ruleRepo, _, _, _, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{})

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	fs.AssertCalled(t, "ListFiles", mock.Anything, "/Downloads", false)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-9: Processor builds correct ProcessingContext before rendering templates
// ─────────────────────────────────────────────────────────────────────────────

// S-9-1: ProcessingContext 的 project 來自規則設定
func TestBackgroundProcessor_ProcessingContext_ProjectFromRule(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	var capturedJob *job.ProcessingJob
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).
		Run(func(args mock.Arguments) { capturedJob = args.Get(1).(*job.ProcessingJob) }).
		Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)
	seqLogRepo.On("CountSuccessByRuleAndDate", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(0, nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	assert.Equal(t, "invoice-project", capturedJob.Context().Project)
}

// S-9-2: ProcessingContext 的 seq 為 "001"（第一筆）
func TestBackgroundProcessor_ProcessingContext_SeqIsFirstValue(t *testing.T) {
	// Arrange
	svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs := createBackgroundProcessorService()
	givenEnabledRuleWithPdfFilter(ruleRepo)
	givenFilesInFolder(fs, "/Downloads", false, []application.FileInfo{createPdfFileInfo()})
	fs.On("EnsureDir", mock.Anything, mock.AnythingOfType("string")).Return(nil)
	fs.On("MoveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	var capturedJob *job.ProcessingJob
	jobRepo.On("Save", mock.Anything, mock.AnythingOfType("*job.ProcessingJob")).
		Run(func(args mock.Arguments) { capturedJob = args.Get(1).(*job.ProcessingJob) }).
		Return(nil)
	logWriter.On("Save", mock.Anything, mock.AnythingOfType("*query.OperationLog")).Return(nil)
	seqLogRepo.On("CountSuccessByRuleAndDate", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(0, nil)

	// Act
	_ = svc.ScanAndProcess(context.Background())

	// Assert
	assert.Equal(t, "001", capturedJob.Context().Seq)
}

// ─────────────────────────────────────────────────────────────────────────────
// Mocks
// ─────────────────────────────────────────────────────────────────────────────

type MockRuleRepo struct{ mock.Mock }

func (m *MockRuleRepo) Save(ctx context.Context, r *rule.Rule) error {
	return m.Called(ctx, r).Error(0)
}
func (m *MockRuleRepo) FindById(ctx context.Context, id rule.RuleId) (*rule.Rule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rule.Rule), args.Error(1)
}
func (m *MockRuleRepo) FindByFolderPath(ctx context.Context, folderPath string) (*rule.Rule, error) {
	args := m.Called(ctx, folderPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rule.Rule), args.Error(1)
}
func (m *MockRuleRepo) ExistsByFolderPath(ctx context.Context, folderPath string) (bool, error) {
	args := m.Called(ctx, folderPath)
	return args.Bool(0), args.Error(1)
}
func (m *MockRuleRepo) Delete(ctx context.Context, id rule.RuleId) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRuleRepo) ListAll(ctx context.Context) ([]*rule.Rule, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*rule.Rule), args.Error(1)
}

type MockJobRepo struct{ mock.Mock }

func (m *MockJobRepo) Save(ctx context.Context, j *job.ProcessingJob) error {
	return m.Called(ctx, j).Error(0)
}
func (m *MockJobRepo) FindById(ctx context.Context, id job.JobId) (*job.ProcessingJob, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*job.ProcessingJob), args.Error(1)
}

type MockLogWriter struct{ mock.Mock }

func (m *MockLogWriter) Save(ctx context.Context, log *query.OperationLog) error {
	return m.Called(ctx, log).Error(0)
}

type MockSeqLogRepo struct{ mock.Mock }

func (m *MockSeqLogRepo) CountSuccessByRuleAndDate(ctx context.Context, ruleId string, date time.Time) (int, error) {
	args := m.Called(ctx, ruleId, date)
	return args.Int(0), args.Error(1)
}

type MockFileSystem struct{ mock.Mock }

func (m *MockFileSystem) ListFiles(ctx context.Context, dir string, recursive bool) ([]application.FileInfo, error) {
	args := m.Called(ctx, dir, recursive)
	return args.Get(0).([]application.FileInfo), args.Error(1)
}
func (m *MockFileSystem) MoveFile(ctx context.Context, src, dst string) error {
	return m.Called(ctx, src, dst).Error(0)
}
func (m *MockFileSystem) EnsureDir(ctx context.Context, dir string) error {
	return m.Called(ctx, dir).Error(0)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createBackgroundProcessorService() (
	*application.BackgroundProcessorService,
	*MockRuleRepo,
	*MockSeqLogRepo,
	*MockJobRepo,
	*MockLogWriter,
	*MockFileSystem,
) {
	ruleRepo := &MockRuleRepo{}
	seqLogRepo := &MockSeqLogRepo{}
	jobRepo := &MockJobRepo{}
	logWriter := &MockLogWriter{}
	fs := &MockFileSystem{}

	svc := application.NewBackgroundProcessorService(
		ruleRepo,
		jobRepo,
		logWriter,
		fs,
		seqLogRepo,
	)
	return svc, ruleRepo, seqLogRepo, jobRepo, logWriter, fs
}

func givenEnabledRuleWithPdfFilter(repo *MockRuleRepo) {
	r := rule.NewRule(
		"發票規則",
		rule.WatchConfig{FolderPath: "/Downloads", Recursive: false},
		rule.FilterSpec{Extensions: []string{".pdf"}, Keyword: ""},
		rule.NamingTemplate{TemplateString: "{project}_{type}_{YYYY}{MM}{DD}_{seq}{ext}"},
		rule.TargetPathTemplate{PathTemplate: "/Documents/{project}/{type}/{YYYY}/{MM}"},
		"invoice-project",
		"invoice",
	)
	repo.On("ListAll", mock.Anything).Return([]*rule.Rule{r}, nil)
}

func givenEnabledRuleWithKeyword(repo *MockRuleRepo, keyword string) {
	r := rule.NewRule(
		"合約規則",
		rule.WatchConfig{FolderPath: "/Downloads", Recursive: false},
		rule.FilterSpec{Extensions: []string{".pdf"}, Keyword: keyword},
		rule.NamingTemplate{TemplateString: "{project}_{type}_{YYYY}{MM}{DD}_{seq}{ext}"},
		rule.TargetPathTemplate{PathTemplate: "/Documents/{project}/{type}/{YYYY}/{MM}"},
		"contract-project",
		"contract",
	)
	repo.On("ListAll", mock.Anything).Return([]*rule.Rule{r}, nil)
}

func givenEnabledRecursiveRule(repo *MockRuleRepo) {
	r := rule.NewRule(
		"發票規則",
		rule.WatchConfig{FolderPath: "/Downloads", Recursive: true},
		rule.FilterSpec{Extensions: []string{".pdf"}, Keyword: ""},
		rule.NamingTemplate{TemplateString: "{project}_{type}_{YYYY}{MM}{DD}_{seq}{ext}"},
		rule.TargetPathTemplate{PathTemplate: "/Documents/{project}/{type}/{YYYY}/{MM}"},
		"invoice-project",
		"invoice",
	)
	repo.On("ListAll", mock.Anything).Return([]*rule.Rule{r}, nil)
}

func givenNoEnabledRules(repo *MockRuleRepo) {
	repo.On("ListAll", mock.Anything).Return([]*rule.Rule{}, nil)
}

func givenSeqIsFirst(repo *MockSeqLogRepo) {
	repo.On("CountSuccessByRuleAndDate", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
		Return(0, nil)
}

func givenFilesInFolder(fs *MockFileSystem, dir string, recursive bool, files []application.FileInfo) {
	fs.On("ListFiles", mock.Anything, dir, recursive).Return(files, nil)
}

func createPdfFileInfo() application.FileInfo {
	return application.FileInfo{
		Path:       "/Downloads/invoice.pdf",
		Name:       "invoice",
		Extension:  ".pdf",
		DetectedAt: time.Now(),
	}
}

func createJpgFileInfo() application.FileInfo {
	return application.FileInfo{
		Path:       "/Downloads/photo.jpg",
		Name:       "photo",
		Extension:  ".jpg",
		DetectedAt: time.Now(),
	}
}

func createPdfSubdirFileInfo() application.FileInfo {
	return application.FileInfo{
		Path:       "/Downloads/2024/invoice_sub.pdf",
		Name:       "invoice_sub",
		Extension:  ".pdf",
		DetectedAt: time.Now(),
	}
}
