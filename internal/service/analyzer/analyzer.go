package analyzer

import (
	"config-analyzer/internal/domain/model"
	"config-analyzer/internal/domain/rules"
	"config-analyzer/internal/infrastructure/config"
	"config-analyzer/internal/infrastructure/output"
	"fmt"
)

type Rule interface {
	Check(*config.Config) *model.Issue
}

type Analyzer struct {
	Rules []Rule
}

func New() *Analyzer {
	return new(Analyzer{
		Rules: []Rule{
			rules.TLSRule{},
			rules.LogRule{},
			rules.StorageRule{},
			rules.PasswordRule{},
			rules.ServerRule{},
		},
	})
}

func (analyzer *Analyzer) Analyze(cfg *config.Config) []model.Issue {
	var issues []model.Issue
	for _, rule := range analyzer.Rules {
		if issue := rule.Check(cfg); issue != nil {
			issues = append(issues, *issue)
		}
	}
	return issues
}


func AnalyzeAndPrint(cfg *config.Config, isSilent bool) error {
	anlzr := New()
	issues := anlzr.Analyze(cfg)

	if len(issues) > 0 {
		output.PrintSortedIssues(issues)
		if !isSilent {
			return fmt.Errorf("issues found")
		}
		return nil
	}

	fmt.Println("No issues found")
	return nil
}
