package rule

import (
	"strings"

	"config-analyzer/internal/model"
)

type PasswordRule struct{}

func (r PasswordRule) Check(cfg *model.Config) *model.Issue {
	password := strings.TrimSpace(cfg.Auth.Password)

	if password != "" && isWeak(password) {
		return &model.Issue{
			Severity:       model.HIGH,
			Message:        "слабый пароль в конфиге",
			Recommendation: "используйте надёжный пароль или secret manager",
			Path:           "auth.password",
		}
	}
	return nil
}

func isWeak(p string) bool {
	if len(p) < 8 || p == "password" {
		return true
	}
	return false
}
