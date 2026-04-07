package output

import (
	"fmt"
	"sort"

	"config-analyzer/internal/domain/model"
)

func PrintSortedIssues(issues []model.Issue) {
	sortIssues(issues)
	for _, i := range issues {
		fmt.Printf("%s: %s (%s)\n", i.Severity, i.Message, i.Path)
		fmt.Printf("  → %s\n\n", i.Recommendation)
	}
}

func sortIssues(issues []model.Issue) {
	if len(issues) > 0 {
		sort.Slice(issues, func(i, j int) bool {
			return model.SeverityRank[issues[i].Severity] > model.SeverityRank[issues[j].Severity]
		})
	}
}
