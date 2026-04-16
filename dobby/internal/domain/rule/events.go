package rule

import "time"

type RuleCreated struct {
	RuleId    string
	CreatedAt time.Time
}

type RuleUpdated struct {
	RuleId    string
	UpdatedAt time.Time
}

type RuleDeleted struct {
	RuleId    string
	DeletedAt time.Time
}

type RuleEnabled struct {
	RuleId    string
	EnabledAt time.Time
}

type RuleDisabled struct {
	RuleId     string
	DisabledAt time.Time
}
