package rule

import (
	"fmt"
	"os"

	"config-analyzer/internal/model"
)

type fileInfo interface {
	Mode() os.FileMode
}

type statFunc func(path string) (fileInfo, error)

var stat statFunc = func(path string) (fileInfo, error) {
	return os.Stat(path)
}

func CheckFile(path string) *model.Issue {
	info, err := stat(path)
	if err != nil {
		return nil
	}
	return checkFileFromInfo(path, info)
}

func checkFileFromInfo(path string, info fileInfo) *model.Issue {
	perm := info.Mode().Perm()

	if perm&0002 != 0 {
		return &model.Issue{
			Severity:       model.HIGH,
			Message:        "файл конфигурации доступен всем на запись",
			Recommendation: fmt.Sprintf("ограничьте права (например chmod 600 %s)", path),
			Path:           "file:permissions",
		}
	}

	return nil
}
