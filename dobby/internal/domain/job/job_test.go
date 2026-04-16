package job_test

import (
	"testing"
	"time"

	"github.com/dobby/filemanager/internal/domain/job"
	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-16: 新建立的 ProcessingJob 初始狀態為 pending
// ─────────────────────────────────────────────────────────────────────────────

// S-16-1: 初始狀態為 pending
func TestNewProcessingJob_InitialStateIsPending(t *testing.T) {
	// Arrange
	event := createFileEvent("~/Downloads/Untitled.png")

	// Act
	j := job.NewProcessingJob(event)

	// Assert
	assert.Equal(t, job.JobStatePending, j.State())
}

// S-16-2: 擁有非空的 JobId
func TestNewProcessingJob_HasNonEmptyJobId(t *testing.T) {
	// Arrange
	event := createFileEvent("~/Downloads/Untitled.png")

	// Act
	j := job.NewProcessingJob(event)

	// Assert
	assert.NotEmpty(t, j.Id().String())
}

// S-16-3: 兩個不同 Job 的 JobId 不相同（唯一性）
func TestNewProcessingJob_IdsAreUnique(t *testing.T) {
	// Arrange
	event := createFileEvent("~/Downloads/Untitled.png")

	// Act
	j1 := job.NewProcessingJob(event)
	j2 := job.NewProcessingJob(event)

	// Assert
	assert.NotEqual(t, j1.Id().String(), j2.Id().String())
}

// ─────────────────────────────────────────────────────────────────────────────
// S-17: RuleMatcher 比對成功後，Job 狀態轉移至 matched
// ─────────────────────────────────────────────────────────────────────────────

// S-17-1: MarkMatched() 後狀態為 matched
func TestProcessingJob_MarkMatched_StateIsMatched(t *testing.T) {
	// Arrange
	j := givenPendingJob()
	ruleId := rule.RuleIdFrom("rule-001")
	ctx := createProcessingContext()

	// Act
	_ = j.MarkMatched(ruleId, ctx)

	// Assert
	assert.Equal(t, job.JobStateMatched, j.State())
}

// S-17-2: MarkMatched() 不回傳錯誤（合法轉移）
func TestProcessingJob_MarkMatched_ReturnsNoError(t *testing.T) {
	// Arrange
	j := givenPendingJob()
	ruleId := rule.RuleIdFrom("rule-001")
	ctx := createProcessingContext()

	// Act
	err := j.MarkMatched(ruleId, ctx)

	// Assert
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-18: 開始執行時，Job 狀態轉移至 processing
// ─────────────────────────────────────────────────────────────────────────────

// S-18-1: MarkProcessing() 後狀態為 processing
func TestProcessingJob_MarkProcessing_StateIsProcessing(t *testing.T) {
	// Arrange
	j := givenMatchedJob()

	// Act
	_ = j.MarkProcessing()

	// Assert
	assert.Equal(t, job.JobStateProcessing, j.State())
}

// S-18-2: MarkProcessing() 不回傳錯誤（合法轉移）
func TestProcessingJob_MarkProcessing_ReturnsNoError(t *testing.T) {
	// Arrange
	j := givenMatchedJob()

	// Act
	err := j.MarkProcessing()

	// Assert
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-19: 處理成功後，Job 狀態轉移至 succeeded
// ─────────────────────────────────────────────────────────────────────────────

const succeededNewPath = "~/Projects/my-app/assets/my-app-screenshot-20260416-001.png"

// S-19-1: MarkSucceeded() 後狀態為 succeeded
func TestProcessingJob_MarkSucceeded_StateIsSucceeded(t *testing.T) {
	// Arrange
	j := givenProcessingJob()

	// Act
	_ = j.MarkSucceeded(succeededNewPath)

	// Assert
	assert.Equal(t, job.JobStateSucceeded, j.State())
}

// S-19-2: MarkSucceeded() 後 Result.NewPath 正確
func TestProcessingJob_MarkSucceeded_NewPathIsSet(t *testing.T) {
	// Arrange
	j := givenProcessingJob()

	// Act
	_ = j.MarkSucceeded(succeededNewPath)

	// Assert
	assert.Equal(t, succeededNewPath, j.Result().NewPath)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-20: 處理失敗後，Job 狀態轉移至 failed
// ─────────────────────────────────────────────────────────────────────────────

const failureMsg = "permission denied"

// S-20-1: MarkFailed() 後狀態為 failed
func TestProcessingJob_MarkFailed_StateIsFailed(t *testing.T) {
	// Arrange
	j := givenProcessingJob()

	// Act
	_ = j.MarkFailed(failureMsg)

	// Assert
	assert.Equal(t, job.JobStateFailed, j.State())
}

// S-20-2: MarkFailed() 後 Result.ErrorMessage 正確
func TestProcessingJob_MarkFailed_ErrorMessageIsSet(t *testing.T) {
	// Arrange
	j := givenProcessingJob()

	// Act
	_ = j.MarkFailed(failureMsg)

	// Assert
	assert.Equal(t, failureMsg, j.Result().ErrorMessage)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-21: 非法狀態轉移應回傳 ErrInvalidStateTransition
// ─────────────────────────────────────────────────────────────────────────────

// S-21-1: 從 pending 直接呼叫 MarkSucceeded 回傳錯誤
func TestProcessingJob_InvalidTransition_PendingToSucceeded(t *testing.T) {
	// Arrange
	j := givenPendingJob()

	// Act
	err := j.MarkSucceeded("~/some/path")

	// Assert
	assert.ErrorIs(t, err, job.ErrInvalidStateTransition)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createFileEvent(path string) job.FileEvent {
	return job.FileEvent{
		DetectedPath: path,
		OriginalName: "Untitled.png",
		Extension:    ".png",
		DetectedAt:   time.Now(),
	}
}

func createProcessingContext() job.ProcessingContext {
	return job.ProcessingContext{
		Project:      "my-app",
		TypeLabel:    "screenshot",
		Date:         time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
		Seq:          "001",
		OriginalName: "Untitled",
		Extension:    "png",
	}
}

func givenPendingJob() *job.ProcessingJob {
	return job.NewProcessingJob(createFileEvent("~/Downloads/Untitled.png"))
}

func givenMatchedJob() *job.ProcessingJob {
	j := givenPendingJob()
	_ = j.MarkMatched(rule.RuleIdFrom("rule-001"), createProcessingContext())
	return j
}

func givenProcessingJob() *job.ProcessingJob {
	j := givenMatchedJob()
	_ = j.MarkProcessing()
	return j
}
