package rule

import "github.com/google/uuid"

// RuleId is a Value Object wrapping a unique string identifier.
type RuleId struct {
	value string
}

func NewRuleId() RuleId {
	return RuleId{value: uuid.NewString()}
}

func RuleIdFrom(s string) RuleId {
	return RuleId{value: s}
}

func (id RuleId) String() string {
	return id.value
}

func (id RuleId) Equals(other RuleId) bool {
	return id.value == other.value
}

// WatchConfig holds the folder monitoring configuration.
type WatchConfig struct {
	FolderPath string
	Recursive  bool
}

// FilterSpec defines which files a rule applies to.
type FilterSpec struct {
	Extensions []string // e.g. [".png", ".jpg"]; empty = accept all
	Keyword    string   // original filename must contain this; empty = no filter
}

// NamingTemplate holds the template string for naming output files.
type NamingTemplate struct {
	TemplateString string
}

// TargetPathTemplate holds the template string for the target folder path.
type TargetPathTemplate struct {
	PathTemplate string
}
