package rules

import (
	"strings"

	"config-analyzer/internal/domain/model"
	"config-analyzer/internal/infrastructure/config"
)

type LogRule struct{}

func (r LogRule) Check(cfg *config.Config) *model.Issue {
	level := strings.ToLower(strings.TrimSpace(cfg.Log.Level))

	if level == "debug" {
		return &model.Issue{
			Severity:       model.LOW,
			Message:        "логирование в debug-режиме",
			Recommendation: "используйте уровень info или выше",
			Path:           "log.level",
		}
	}
	return nil
}
