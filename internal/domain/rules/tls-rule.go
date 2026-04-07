package rules

import (
	"config-analyzer/internal/domain/model"
	"config-analyzer/internal/infrastructure/config"
)

type TLSRule struct{}

func (r TLSRule) Check(cfg *config.Config) *model.Issue {
	if !cfg.TLS.Enabled {
		return &model.Issue{
			Severity:       model.HIGH,
			Message:        "TLS отключен",
			Recommendation: "включите TLS для защиты соединения",
			Path:           "tls.enabled",
		}
	}
	return nil
}
