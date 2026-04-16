package service

import (
	"strings"

	"github.com/dobby/filemanager/internal/domain/rule"
)

// RuleMatcher is a Domain Service that decides whether a FileEvent matches a Rule's FilterSpec.
type RuleMatcher struct{}

func NewRuleMatcher() *RuleMatcher {
	return &RuleMatcher{}
}

// Match returns true when the fileEvent satisfies all FilterSpec conditions (AND logic).
// - If Extensions is empty, all extensions are accepted.
// - If Keyword is empty, no keyword filtering is applied.
func (m *RuleMatcher) Match(extension string, fileName string, spec rule.FilterSpec) bool {
	if len(spec.Extensions) > 0 {
		found := false
		for _, ext := range spec.Extensions {
			if ext == extension {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if spec.Keyword != "" && !strings.Contains(fileName, spec.Keyword) {
		return false
	}

	return true
}
