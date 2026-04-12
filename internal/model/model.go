package model

type Severity string

const (
	LOW    Severity = "LOW"
	MEDIUM Severity = "MEDIUM"
	HIGH   Severity = "HIGH"
)

type Issue struct {
	Severity       Severity `json:"severity"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation,omitempty"`
	Path           string   `json:"path"`
}

type Report struct {
	File   string
	Issues []Issue
	Err    error
}

type Summary struct {
	FilesWithIssues int
	TotalIssues     int
}

type Config struct {
	Version string  `json:"version,omitempty" yaml:"version,omitempty"`
	Server  Server  `json:"server" yaml:"server"`
	Auth    Auth    `json:"auth" yaml:"auth"`
	Log     Log     `json:"log" yaml:"log"`
	TLS     TLS     `json:"tls" yaml:"tls"`
	Storage Storage `json:"storage" yaml:"storage"`
}

type Server struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Port int    `json:"port,omitempty" yaml:"port"`
}

type Auth struct {
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

type TLS struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Log struct {
	Output string `json:"output,omitempty" yaml:"output,omitempty"`
	Level  string `json:"level,omitempty" yaml:"level,omitempty"`
}

type Storage struct {
	DigestAlgorithm string `json:"digest-algorithm,omitempty" yaml:"digest-algorithm,omitempty"`
}
