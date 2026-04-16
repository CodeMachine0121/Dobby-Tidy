package job

import "time"

type FileDetected struct {
	JobId      string
	FilePath   string
	DetectedAt time.Time
}

type RuleMatched struct {
	JobId     string
	RuleId    string
	MatchedAt time.Time
}

type ProcessingSucceeded struct {
	JobId       string
	NewPath     string
	ProcessedAt time.Time
}

type ProcessingFailed struct {
	JobId        string
	ErrorMessage string
	FailedAt     time.Time
}
