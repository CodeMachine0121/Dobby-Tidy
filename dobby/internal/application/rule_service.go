package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dobby/filemanager/internal/domain/rule"
)

var ErrDuplicateWatchFolder = errors.New("watch folder is already monitored by another rule")
var ErrRuleNotFound = errors.New("rule not found")

// CreateRuleCmd is the input for creating a new rule.
type CreateRuleCmd struct {
	Name           string
	WatchFolder    string
	Recursive      bool
	FilterExts     []string
	FilterKeyword  string
	NameTemplate   string
	TargetTemplate string
	Project        string
	TypeLabel      string
}

// RuleService orchestrates Rule use-cases.
type RuleService struct {
	repo rule.IRuleRepository
}

func NewRuleService(repo rule.IRuleRepository) *RuleService {
	return &RuleService{repo: repo}
}

// CreateRule validates uniqueness, builds the aggregate, and persists it.
func (s *RuleService) CreateRule(ctx context.Context, cmd CreateRuleCmd) (*rule.Rule, error) {
	exists, err := s.repo.ExistsByFolderPath(ctx, cmd.WatchFolder)
	if err != nil {
		return nil, fmt.Errorf("rule_service.CreateRule check: %w", err)
	}
	if exists {
		return nil, ErrDuplicateWatchFolder
	}

	r := rule.NewRule(
		cmd.Name,
		rule.WatchConfig{FolderPath: cmd.WatchFolder, Recursive: cmd.Recursive},
		rule.FilterSpec{Extensions: cmd.FilterExts, Keyword: cmd.FilterKeyword},
		rule.NamingTemplate{TemplateString: cmd.NameTemplate},
		rule.TargetPathTemplate{PathTemplate: cmd.TargetTemplate},
		cmd.Project,
		cmd.TypeLabel,
	)
	if err := s.repo.Save(ctx, r); err != nil {
		return nil, fmt.Errorf("rule_service.CreateRule save: %w", err)
	}
	return r, nil
}

// EnableRule loads, enables, and saves the rule.
func (s *RuleService) EnableRule(ctx context.Context, id string) error {
	r, err := s.findOrErr(ctx, id)
	if err != nil {
		return err
	}
	if err := r.Enable(); err != nil {
		return err
	}
	return s.repo.Save(ctx, r)
}

// DisableRule loads, disables, and saves the rule.
func (s *RuleService) DisableRule(ctx context.Context, id string) error {
	r, err := s.findOrErr(ctx, id)
	if err != nil {
		return err
	}
	if err := r.Disable(); err != nil {
		return err
	}
	return s.repo.Save(ctx, r)
}

// DeleteRule removes the rule from the repository.
func (s *RuleService) DeleteRule(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, rule.RuleIdFrom(id))
}

// GetRule retrieves a single rule.
func (s *RuleService) GetRule(ctx context.Context, id string) (*rule.Rule, error) {
	return s.findOrErr(ctx, id)
}

// ListRules returns all rules.
func (s *RuleService) ListRules(ctx context.Context) ([]*rule.Rule, error) {
	return s.repo.ListAll(ctx)
}

func (s *RuleService) findOrErr(ctx context.Context, id string) (*rule.Rule, error) {
	r, err := s.repo.FindById(ctx, rule.RuleIdFrom(id))
	if err != nil {
		return nil, fmt.Errorf("rule_service find: %w", err)
	}
	if r == nil {
		return nil, ErrRuleNotFound
	}
	return r, nil
}
