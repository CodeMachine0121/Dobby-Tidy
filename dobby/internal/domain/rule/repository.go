package rule

import "context"

// IRuleRepository is the persistence contract for the Rule aggregate.
// Infrastructure implements this; domain declares it (Dependency Inversion).
type IRuleRepository interface {
	Save(ctx context.Context, rule *Rule) error
	FindById(ctx context.Context, id RuleId) (*Rule, error)
	FindByFolderPath(ctx context.Context, folderPath string) (*Rule, error)
	ExistsByFolderPath(ctx context.Context, folderPath string) (bool, error)
	Delete(ctx context.Context, id RuleId) error
	ListAll(ctx context.Context) ([]*Rule, error)
}
