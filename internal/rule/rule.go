package rule

import "config-analyzer/internal/model"


type Rule interface {
	Check(cfg *model.Config) *model.Issue
}

func NewDefaultRules() []Rule {
	return []Rule{
		TLSRule{},
		LogRule{},
		StorageRule{},
		PasswordRule{},
		ServerRule{},
	}
}
