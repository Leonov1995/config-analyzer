package rule

import (
	"strings"

	"config-analyzer/internal/model"
)

type LogRule struct{}

func (r LogRule) Check(cfg *model.Config) *model.Issue {
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
