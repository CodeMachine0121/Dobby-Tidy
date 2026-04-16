package rule

import (
	"errors"
	"time"
)

var ErrAlreadyEnabled = errors.New("rule is already enabled")
var ErrAlreadyDisabled = errors.New("rule is already disabled")

// Rule is the Aggregate Root for file-processing rule configuration.
type Rule struct {
	id             RuleId
	name           string
	enabled        bool
	watchConfig    WatchConfig
	filterSpec     FilterSpec
	nameTemplate   NamingTemplate
	targetTemplate TargetPathTemplate
	project        string
	typeLabel      string
	createdAt      time.Time
	updatedAt      time.Time
}

// NewRule constructs a Rule. All new rules are enabled by default.
func NewRule(
	name string,
	watchConfig WatchConfig,
	filterSpec FilterSpec,
	nameTemplate NamingTemplate,
	targetTemplate TargetPathTemplate,
	project string,
	typeLabel string,
) *Rule {
	now := time.Now()
	return &Rule{
		id:             NewRuleId(),
		name:           name,
		enabled:        true,
		watchConfig:    watchConfig,
		filterSpec:     filterSpec,
		nameTemplate:   nameTemplate,
		targetTemplate: targetTemplate,
		project:        project,
		typeLabel:      typeLabel,
		createdAt:      now,
		updatedAt:      now,
	}
}

// Accessors
func (r *Rule) Id() RuleId                         { return r.id }
func (r *Rule) Name() string                       { return r.name }
func (r *Rule) Enabled() bool                      { return r.enabled }
func (r *Rule) WatchConfig() WatchConfig           { return r.watchConfig }
func (r *Rule) FilterSpec() FilterSpec             { return r.filterSpec }
func (r *Rule) NameTemplate() NamingTemplate       { return r.nameTemplate }
func (r *Rule) TargetTemplate() TargetPathTemplate { return r.targetTemplate }
func (r *Rule) Project() string                    { return r.project }
func (r *Rule) TypeLabel() string                  { return r.typeLabel }
func (r *Rule) CreatedAt() time.Time               { return r.createdAt }
func (r *Rule) UpdatedAt() time.Time               { return r.updatedAt }

// Enable activates the rule.
func (r *Rule) Enable() error {
	if r.enabled {
		return ErrAlreadyEnabled
	}
	r.enabled = true
	r.updatedAt = time.Now()
	return nil
}

// Disable deactivates the rule.
func (r *Rule) Disable() error {
	if !r.enabled {
		return ErrAlreadyDisabled
	}
	r.enabled = false
	r.updatedAt = time.Now()
	return nil
}
