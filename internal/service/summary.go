package service

import "config-analyzer/internal/model"

func buildSummary(reports []model.Report) *model.Summary {
	sum := new(model.Summary)

	for _, r := range reports {
		if len(r.Issues) > 0 {
			sum.FilesWithIssues++
			sum.TotalIssues += len(r.Issues)
		}
	}

	return sum
}
