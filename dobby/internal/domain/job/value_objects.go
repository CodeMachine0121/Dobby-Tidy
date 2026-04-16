package job

import (
	"github.com/google/uuid"
	"time"
)

// JobId is a Value Object wrapping a unique string identifier.
type JobId struct {
	value string
}

func NewJobId() JobId {
	return JobId{value: uuid.NewString()}
}

func JobIdFrom(s string) JobId {
	return JobId{value: s}
}

func (id JobId) String() string {
	return id.value
}

// FileEvent captures the detected file information.
type FileEvent struct {
	DetectedPath string
	OriginalName string
	Extension    string
	DetectedAt   time.Time
}

// JobState represents the lifecycle state of a ProcessingJob.
type JobState string

const (
	JobStatePending    JobState = "pending"
	JobStateMatched    JobState = "matched"
	JobStateProcessing JobState = "processing"
	JobStateSucceeded  JobState = "succeeded"
	JobStateFailed     JobState = "failed"
)

// ProcessingContext holds the resolved variable values used by TemplateRenderer.
type ProcessingContext struct {
	Project      string
	TypeLabel    string
	Date         time.Time
	Seq          string
	OriginalName string
	Extension    string
}

// ProcessingResult records the outcome of a processing attempt.
type ProcessingResult struct {
	NewPath      string
	ErrorMessage string
	ProcessedAt  time.Time
}
