package config

import (
	"bytes"
	"encoding/json"

	"fmt"
	"io"
	"os"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	Version string  `mapstructure:"version"`
	Server  Server  `mapstructure:"server"`
	Auth    Auth    `mapstructure:"auth"`
	Log     Log     `mapstructure:"log"`
	TLS     TLS     `mapstructure:"tls"`
	Storage Storage `mapstructure:"storage"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Auth struct {
	Password string `mapstructure:"password"`
}

type TLS struct {
	Enabled bool `mapstructure:"enabled"`
}

type Log struct {
	Output string `mapstructure:"output"`
	Level  string `mapstructure:"level"`
}

type Storage struct {
	DigestAlgorithm string `mapstructure:"digest-algorithm" yaml:"digest-algorithm" json:"digest-algorithm"`
}

func Load(fileName string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(fileName)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", fileName, err)
	}

	cfg := new(Config)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func ParseDataWithOrWithoutFormat(data []byte, format string) (*Config, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("empty config data")
	}
	fmt.Println(string(data))

	cfg := new(Config)

	switch format {
	case "json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	case "":
		if err := json.Unmarshal(data, cfg); err == nil {
			return cfg, nil
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config as JSON or YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid format: %s, must be 'json' or 'yaml'", format)
	}
	fmt.Println(cfg)

	return cfg, nil
}

func LoadConfig(args []string, isSTDIN bool, format string) (*Config, error) {
	cfg := new(Config)
	var err error

	if isSTDIN {
		fmt.Println("ParseData")
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
		cfg, err = ParseDataWithOrWithoutFormat(data, format)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config from stdin: %w", err)
		}
	} else {
		fmt.Println("Load")
		if len(args) == 0 {
			return nil, fmt.Errorf("usage: config-analyzer [file]")
		}
		cfg, err = Load(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to load config '%s': %w", args[0], err)
		}
	}

	return cfg, nil
}
