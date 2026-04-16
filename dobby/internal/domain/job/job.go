package job

import (
	"errors"
	"time"

	"github.com/dobby/filemanager/internal/domain/rule"
)

var ErrInvalidStateTransition = errors.New("invalid state transition")

// ProcessingJob is the Aggregate Root that tracks the lifecycle of a file-processing operation.
type ProcessingJob struct {
	id        JobId
	ruleId    rule.RuleId
	fileEvent FileEvent
	state     JobState
	context   *ProcessingContext
	result    *ProcessingResult
}

// NewProcessingJob creates a new job in the pending state.
func NewProcessingJob(event FileEvent) *ProcessingJob {
	return &ProcessingJob{
		id:        NewJobId(),
		fileEvent: event,
		state:     JobStatePending,
	}
}

// Accessors
func (j *ProcessingJob) Id() JobId                   { return j.id }
func (j *ProcessingJob) RuleId() rule.RuleId         { return j.ruleId }
func (j *ProcessingJob) FileEvent() FileEvent        { return j.fileEvent }
func (j *ProcessingJob) State() JobState             { return j.state }
func (j *ProcessingJob) Context() *ProcessingContext { return j.context }
func (j *ProcessingJob) Result() *ProcessingResult   { return j.result }

// MarkMatched transitions from pending → matched and attaches the matched rule + context.
func (j *ProcessingJob) MarkMatched(ruleId rule.RuleId, ctx ProcessingContext) error {
	if j.state != JobStatePending {
		return ErrInvalidStateTransition
	}
	j.ruleId = ruleId
	j.context = &ctx
	j.state = JobStateMatched
	return nil
}

// MarkProcessing transitions from matched → processing.
func (j *ProcessingJob) MarkProcessing() error {
	if j.state != JobStateMatched {
		return ErrInvalidStateTransition
	}
	j.state = JobStateProcessing
	return nil
}

// MarkSucceeded transitions from processing → succeeded and records the new file path.
func (j *ProcessingJob) MarkSucceeded(newPath string) error {
	if j.state != JobStateProcessing {
		return ErrInvalidStateTransition
	}
	j.result = &ProcessingResult{
		NewPath:     newPath,
		ProcessedAt: time.Now(),
	}
	j.state = JobStateSucceeded
	return nil
}

// MarkFailed transitions from processing → failed and records the error message.
func (j *ProcessingJob) MarkFailed(errorMessage string) error {
	if j.state != JobStateProcessing {
		return ErrInvalidStateTransition
	}
	j.result = &ProcessingResult{
		ErrorMessage: errorMessage,
		ProcessedAt:  time.Now(),
	}
	j.state = JobStateFailed
	return nil
}

// Reconstitute rebuilds a ProcessingJob from persisted data (used by repositories only).
func Reconstitute(
	id JobId,
	ruleId rule.RuleId,
	fileEvent FileEvent,
	state JobState,
	ctx *ProcessingContext,
	result *ProcessingResult,
) *ProcessingJob {
	return &ProcessingJob{
		id:        id,
		ruleId:    ruleId,
		fileEvent: fileEvent,
		state:     state,
		context:   ctx,
		result:    result,
	}
}
