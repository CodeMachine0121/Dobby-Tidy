package rule_test

import (
	"testing"

	"github.com/dobby/filemanager/internal/domain/rule"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// S-1: 成功建立一條完整規則
// ─────────────────────────────────────────────────────────────────────────────

// S-1-1: 新建立的 Rule enabled 狀態為 true
func TestNewRule_EnabledByDefault(t *testing.T) {
	// Arrange
	r := createDefaultRule()

	// Act
	enabled := r.Enabled()

	// Assert
	assert.True(t, enabled)
}

// S-1-2: 新建立的 Rule 擁有非空的 RuleId
func TestNewRule_HasNonEmptyRuleId(t *testing.T) {
	// Arrange
	r := createDefaultRule()

	// Act
	id := r.Id().String()

	// Assert
	assert.NotEmpty(t, id)
}

// S-1-3: 兩個不同 Rule 的 RuleId 不相同（唯一性）
func TestNewRule_IdsAreUnique(t *testing.T) {
	// Arrange
	r1 := createDefaultRule()
	r2 := createDefaultRule()

	// Act — compare IDs directly
	id1 := r1.Id().String()
	id2 := r2.Id().String()

	// Assert
	assert.NotEqual(t, id1, id2)
}

// ─────────────────────────────────────────────────────────────────────────────
// S-3: 停用一條 enabled 的規則
// ─────────────────────────────────────────────────────────────────────────────

// S-3-1: Disable() 不回傳錯誤
func TestRule_Disable_ReturnsNoError(t *testing.T) {
	// Arrange
	r := givenEnabledRule()

	// Act
	err := r.Disable()

	// Assert
	assert.NoError(t, err)
}

// S-3-2: Disable() 後 Enabled() 為 false
func TestRule_Disable_SetsEnabledFalse(t *testing.T) {
	// Arrange
	r := givenEnabledRule()

	// Act
	_ = r.Disable()

	// Assert
	assert.False(t, r.Enabled())
}

// ─────────────────────────────────────────────────────────────────────────────
// S-4: 啟用一條已停用的規則
// ─────────────────────────────────────────────────────────────────────────────

// S-4-1: Enable() 不回傳錯誤
func TestRule_Enable_ReturnsNoError(t *testing.T) {
	// Arrange
	r := givenDisabledRule()

	// Act
	err := r.Enable()

	// Assert
	assert.NoError(t, err)
}

// S-4-2: Enable() 後 Enabled() 為 true
func TestRule_Enable_SetsEnabledTrue(t *testing.T) {
	// Arrange
	r := givenDisabledRule()

	// Act
	_ = r.Enable()

	// Assert
	assert.True(t, r.Enabled())
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createDefaultRule() *rule.Rule {
	return rule.NewRule(
		"test-rule",
		rule.WatchConfig{FolderPath: "~/Downloads", Recursive: false},
		rule.FilterSpec{Extensions: []string{".png", ".jpg"}, Keyword: ""},
		rule.NamingTemplate{TemplateString: "{project}-{type}-{YYYY}{MM}{DD}-{seq}.{ext}"},
		rule.TargetPathTemplate{PathTemplate: "~/Projects/design/"},
		"my-app",
		"screenshot",
	)
}

func givenEnabledRule() *rule.Rule {
	r := createDefaultRule()
	// Rule is enabled by default; no extra setup needed.
	return r
}

func givenDisabledRule() *rule.Rule {
	r := createDefaultRule()
	_ = r.Disable()
	return r
}
