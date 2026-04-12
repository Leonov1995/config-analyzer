package analyzer

import (
	"config-analyzer/internal/model"
	"config-analyzer/internal/rule"
)

type Analyzer struct {
	rules []rule.Rule
}

func New(rules []rule.Rule) *Analyzer {
	return &Analyzer{rules: rules}
}

func (a *Analyzer) Analyze(cfg *model.Config) []model.Issue {
	if cfg == nil {
		return []model.Issue{
			{Severity: model.HIGH, Message: "nil config"},
		}
	}

	issues := make([]model.Issue, 0, len(a.rules))

	for _, r := range a.rules {
		if issue := r.Check(cfg); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}
