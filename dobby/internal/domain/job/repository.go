package job

import "context"

// IProcessingJobRepository is the persistence contract for ProcessingJob.
type IProcessingJobRepository interface {
	Save(ctx context.Context, job *ProcessingJob) error
	FindById(ctx context.Context, id JobId) (*ProcessingJob, error)
}
