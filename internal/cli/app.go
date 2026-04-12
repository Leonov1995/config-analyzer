package cli

import (
	"config-analyzer/internal/model"
	"fmt"
)

type Service interface {
	Run(stdin bool, args []string, format string) ([]model.Report, *model.Summary, error)
}

type App struct {
	service Service
}

func NewApp(s Service) *App {
	return &App{service: s}
}

func (a *App) Run(stdin bool, args []string, format string, silent bool) error {
	reports, summary, err := a.service.Run(stdin, args, format)
	if err != nil {
		return err
	}
	printResult(reports, summary)

	if !silent && hasIssues(reports) {
		return fmt.Errorf("issues found")
	}
	return nil
}

func hasIssues(reports []model.Report) bool {
	for _, r := range reports {
		if len(r.Issues) > 0 {
			return true
		}
	}
	return false
}
