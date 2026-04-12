package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func collectFiles(path string) ([]string, error) {
	files := make([]string, 0, 5)

	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("path not accessible: %w", err)
	}

	if !info.IsDir() {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("file not readable: %w", err)
		}
		f.Close()

		return []string{path}, nil
	}

	err = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if isConfigFile(p) {
			files = append(files, p)
		}

		return nil
	})

	return files, err
}

func isConfigFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".yaml" || ext == ".yml"
}
