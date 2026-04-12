package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"config-analyzer/internal/model"

	"go.yaml.in/yaml/v3"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func parse(data []byte, format string) (*model.Config, error) {
	cfg := new(model.Config)

	unmarshal := func(d []byte, f string) error {
		switch f {
		case "json":
			return json.Unmarshal(d, cfg)
		case "yaml":
			return yaml.Unmarshal(d, cfg)
		default:
			return nil
		}
	}

	if err := unmarshal(data, format); format != "auto" && err == nil {
		return cfg, nil
	} else if format != "auto" {
		return nil, fmt.Errorf("failed to parse %s: %w", format, err)
	}

	if err := json.Unmarshal(data, cfg); err == nil {
		return cfg, nil
	}
	if err := yaml.Unmarshal(data, cfg); err == nil {
		return cfg, nil
	}
	return nil, fmt.Errorf("failed to parse config as JSON or YAML")
}

func (l *Loader) LoadFromFile(fileName string) (*model.Config, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", fileName, err)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	var format string
	switch ext {
	case ".json":
		format = "json"
	case ".yaml", ".yml":
		format = "yaml"
	default:
		format = "auto"
	}

	cfg, err := parse(data, format)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (l *Loader) ParseDataWithOrWithoutFormat(data []byte, format string) (*model.Config, error) {
	data = bytes.TrimSpace(data)

	if len(data) == 0 {
		return nil, fmt.Errorf("empty config data")
	}

	if format != "" && format != "json" && format != "yaml" {
		return nil, fmt.Errorf("invalid format: %s (must be json|yaml)", format)
	}

	if format == "" {
		format = "auto"
	}

	return parse(data, format)
}

func (l *Loader) LoadFromInput(args []string, isSTDIN bool, format string) (*model.Config, error) {
	if isSTDIN {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
		return l.ParseDataWithOrWithoutFormat(data, format)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("usage: config-analyzer [file]")
	}

	return l.LoadFromFile(args[0])
}
