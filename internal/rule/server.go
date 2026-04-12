package rule

import "config-analyzer/internal/model"

type ServerRule struct{}

func (r ServerRule) Check(cfg *model.Config) *model.Issue {
	s := cfg.Server

	if s.Host == "" {
		return &model.Issue{
			Severity:       model.MEDIUM,
			Message:        "пустой host",
			Recommendation: "укажите host (например 127.0.0.1)",
			Path:           "server.host",
		}
	}

	if s.Host == "0.0.0.0" {
		return &model.Issue{
			Severity:       model.MEDIUM,
			Message:        "приложение слушает на 0.0.0.0",
			Recommendation: "ограничьте интерфейс (например 127.0.0.1)",
			Path:           "server.host",
		}
	}

	return nil
}
