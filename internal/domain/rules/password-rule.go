package rules

import (
	"strings"

	"config-analyzer/internal/domain/model"
	"config-analyzer/internal/infrastructure/config"
)

type PasswordRule struct{}

func (r PasswordRule) Check(cfg *config.Config) *model.Issue {
	if strings.TrimSpace(cfg.Auth.Password) != "" {
		return &model.Issue{
			Severity:       model.HIGH,
			Message:        "пароль в открытом виде",
			Recommendation: "используйте переменные окружения или специальное хранилище секретов",
			Path:           "auth.password",
		}
	}
	return nil
}
