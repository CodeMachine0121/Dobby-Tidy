package service_test

import (
	"testing"

	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/dobby/filemanager/internal/domain/service"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-6: 副檔名符合，無關鍵字限制 → true
// ─────────────────────────────────────────────────────────────────────────────

// S-6-1
func TestRuleMatcher_ExtensionMatch_NoKeyword_ReturnsTrue(t *testing.T) {
	// Arrange
	matcher := service.NewRuleMatcher()
	spec := givenFilterSpecExtensionsOnly([]string{".png", ".jpg"})

	// Act
	result := matcher.Match(".png", "screenshot-001.png", spec)

	// Assert
	assert.True(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-7: 副檔名不符合 → false
// ─────────────────────────────────────────────────────────────────────────────

// S-7-1
func TestRuleMatcher_ExtensionMismatch_ReturnsFalse(t *testing.T) {
	// Arrange
	matcher := service.NewRuleMatcher()
	spec := givenFilterSpecExtensionsOnly([]string{".png", ".jpg"})

	// Act
	result := matcher.Match(".pdf", "report.pdf", spec)

	// Assert
	assert.False(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-8: 副檔名符合 + 關鍵字符合 → true
// ─────────────────────────────────────────────────────────────────────────────

// S-8-1
func TestRuleMatcher_ExtensionAndKeywordMatch_ReturnsTrue(t *testing.T) {
	// Arrange
	matcher := service.NewRuleMatcher()
	spec := givenFilterSpecWithKeyword([]string{".png"}, "export")

	// Act
	result := matcher.Match(".png", "figma-export-v2.png", spec)

	// Assert
	assert.True(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-9: 副檔名符合但關鍵字不符合 → false
// ─────────────────────────────────────────────────────────────────────────────

// S-9-1
func TestRuleMatcher_ExtensionMatchKeywordMismatch_ReturnsFalse(t *testing.T) {
	// Arrange
	matcher := service.NewRuleMatcher()
	spec := givenFilterSpecWithKeyword([]string{".png"}, "export")

	// Act
	result := matcher.Match(".png", "screenshot.png", spec)

	// Assert
	assert.False(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-10: 空副檔名清單時，任何副檔名均通過 → true
// ─────────────────────────────────────────────────────────────────────────────

// S-10-1
func TestRuleMatcher_EmptyExtensions_AcceptsAnyExtension(t *testing.T) {
	// Arrange
	matcher := service.NewRuleMatcher()
	spec := givenFilterSpecExtensionsOnly([]string{})

	// Act
	result := matcher.Match(".pdf", "anything.pdf", spec)

	// Assert
	assert.True(t, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func givenFilterSpecExtensionsOnly(extensions []string) rule.FilterSpec {
	return rule.FilterSpec{Extensions: extensions, Keyword: ""}
}

func givenFilterSpecWithKeyword(extensions []string, keyword string) rule.FilterSpec {
	return rule.FilterSpec{Extensions: extensions, Keyword: keyword}
}
