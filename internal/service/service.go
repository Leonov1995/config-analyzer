package service

import "config-analyzer/internal/model"

type Analyzer interface {
	Analyze(cfg *model.Config) []model.Issue
}

type ConfigLoader interface {
	LoadFromFile(fileName string) (*model.Config, error)
	ParseDataWithOrWithoutFormat(data []byte, format string) (*model.Config, error)
	LoadFromInput(args []string, isSTDIN bool, format string) (*model.Config, error)
}

type Service struct {
	analyzer     Analyzer
	configLoader ConfigLoader
	workers      int
}

func New(a Analyzer, cfgL ConfigLoader, workers int) *Service {
	if workers < 1 {
		workers = 1
	}
	return &Service{
		analyzer:     a,
		configLoader: cfgL,
		workers:      workers,
	}
}
