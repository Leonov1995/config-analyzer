package service

import (
	"config-analyzer/internal/model"
	"config-analyzer/internal/rule"
	"sync"
)

func (s *Service) AnalyzePath(path string) ([]model.Report, *model.Summary, error) {
	files, err := collectFiles(path)
	if err != nil {
		return nil, nil, err
	}

	if len(files) == 0 {
		return nil, nil, nil
	}

	reports := s.analyzeParallel(files)

	return reports, buildSummary(reports), nil
}

func (s *Service) AnalyzeStdin(format string) ([]model.Report, *model.Summary, error) {
	cfg, err := s.configLoader.LoadFromInput(nil, true, format)
	if err != nil {
		return nil, nil, err
	}

	issues := s.analyzer.Analyze(cfg)

	report := model.Report{
		File:   "stdin",
		Issues: issues,
	}

	summary := &model.Summary{
		FilesWithIssues: 0,
		TotalIssues:     len(issues),
	}

	if len(issues) > 0 {
		summary.FilesWithIssues = 1
	}

	return []model.Report{report}, summary, nil
}

func (s *Service) analyzeParallel(files []string) []model.Report {
	reports := make([]model.Report, len(files))
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.workers)

	for i, f := range files {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, file string) {
			defer wg.Done()
			defer func() { <-sem }()

			cfg, err := s.configLoader.LoadFromFile(file)
			if err != nil {
				reports[idx] = model.Report{File: file, Err: err}
				return
			}

			issues := s.analyzer.Analyze(cfg)

			if permIssue := rule.CheckFile(file); permIssue != nil {
				issues = append(issues, *permIssue)
			}

			reports[idx] = model.Report{
				File:   file,
				Issues: issues,
			}
		}(i, f)
	}

	wg.Wait()
	return reports
}
