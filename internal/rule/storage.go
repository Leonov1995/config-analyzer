package rule

import (
	"fmt"
	"strings"

	"config-analyzer/internal/model"
)

type StorageRule struct{}

func (r StorageRule) Check(cfg *model.Config) *model.Issue {
	algo := strings.ToLower(strings.TrimSpace(cfg.Storage.DigestAlgorithm))

	if algo == "md5" || algo == "sha1" {
		return &model.Issue{
			Severity:       model.HIGH,
			Message:        fmt.Sprintf("небезопасный алгоритм: %s", cfg.Storage.DigestAlgorithm),
			Recommendation: "замените алгоритм на более безопасный",
			Path:           "storage.digest-algorithm",
		}
	}
	return nil
}
