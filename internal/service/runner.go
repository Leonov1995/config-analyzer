package service

import (
	"config-analyzer/internal/model"
	"fmt"
)

func (s *Service) Run(stdin bool, args []string, format string) ([]model.Report, *model.Summary, error) {
	if stdin {
		return s.AnalyzeStdin(format)
	}

	if len(args) == 0 {
		return nil, nil, fmt.Errorf("path is required")
	}

	return s.AnalyzePath(args[0])
}
