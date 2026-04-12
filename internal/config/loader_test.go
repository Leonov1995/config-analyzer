package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte(`{
		"version": "1.0",
		"server": {"host": "127.0.0.1", "port": 8080},
		"tls": {"enabled": true},
		"auth": {"password": "supersecret123"},
		"log": {"level": "info", "output": "stdout"},
		"storage": {"digest-algorithm": "sha256"}
	}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	cfg, err := l.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "1.0")
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("server.port = %d, want %d", cfg.Server.Port, 8080)
	}
	if !cfg.TLS.Enabled {
		t.Error("tls.enabled = false, want true")
	}
	if cfg.Auth.Password != "supersecret123" {
		t.Errorf("auth.password = %q, want %q", cfg.Auth.Password, "supersecret123")
	}
}

func TestLoadFromFile_YAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`
version: "2.0"
server:
  host: 0.0.0.0
  port: 3000
tls:
  enabled: false
auth:
  password: "weak"
log:
  level: debug
  output: stderr
storage:
  digest-algorithm: md5
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	cfg, err := l.LoadFromFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != "2.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "2.0")
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.TLS.Enabled {
		t.Error("tls.enabled = true, want false")
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("log.level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.Storage.DigestAlgorithm != "md5" {
		t.Errorf("storage.digest-algorithm = %q, want %q", cfg.Storage.DigestAlgorithm, "md5")
	}

	// .yml extension
	path2 := filepath.Join(dir, "config2.yml")
	err = os.WriteFile(path2, []byte(`version: "3.0"`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cfg2, err := l.LoadFromFile(path2)
	if err != nil {
		t.Fatalf("failed to load .yml file: %v", err)
	}
	if cfg2.Version != "3.0" {
		t.Errorf("version = %q, want %q", cfg2.Version, "3.0")
	}
}

func TestLoadFromFile_AutoDetect(t *testing.T) {
	dir := t.TempDir()

	// JSON without .json extension
	pathJSON := filepath.Join(dir, "config")
	err := os.WriteFile(pathJSON, []byte(`{"version": "auto-json"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	cfg, err := l.LoadFromFile(pathJSON)
	if err != nil {
		t.Fatalf("auto-detect JSON failed: %v", err)
	}
	if cfg.Version != "auto-json" {
		t.Errorf("version = %q, want %q", cfg.Version, "auto-json")
	}

	// YAML without extension
	pathYAML := filepath.Join(dir, "config2")
	err = os.WriteFile(pathYAML, []byte("version: auto-yaml"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	cfg2, err := l.LoadFromFile(pathYAML)
	if err != nil {
		t.Fatalf("auto-detect YAML failed: %v", err)
	}
	if cfg2.Version != "auto-yaml" {
		t.Errorf("version = %q, want %q", cfg2.Version, "auto-yaml")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte(`{invalid json}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	_, err = l.LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(`invalid: yaml: [unterminated`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	_, err = l.LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	l := NewLoader()
	_, err := l.LoadFromFile("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestLoadFromFile_Empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	l := NewLoader()
	_, err = l.LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestParseDataWithOrWithoutFormat_JSON(t *testing.T) {
	l := NewLoader()
	cfg, err := l.ParseDataWithOrWithoutFormat([]byte(`{"version":"1.0","server":{"host":"127.0.0.1"}}`), "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "1.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "1.0")
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
}

func TestParseDataWithOrWithoutFormat_YAML(t *testing.T) {
	l := NewLoader()
	cfg, err := l.ParseDataWithOrWithoutFormat([]byte("version: \"2.0\"\nserver:\n  host: 0.0.0.0"), "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "2.0" {
		t.Errorf("version = %q, want %q", cfg.Version, "2.0")
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
}

func TestParseDataWithOrWithoutFormat_AutoDetect(t *testing.T) {
	l := NewLoader()

	// Auto-detect JSON
	cfg1, err := l.ParseDataWithOrWithoutFormat([]byte(`{"version":"auto1"}`), "")
	if err != nil {
		t.Fatalf("auto JSON failed: %v", err)
	}
	if cfg1.Version != "auto1" {
		t.Errorf("version = %q, want %q", cfg1.Version, "auto1")
	}

	// Auto-detect YAML
	cfg2, err := l.ParseDataWithOrWithoutFormat([]byte("version: auto2"), "")
	if err != nil {
		t.Fatalf("auto YAML failed: %v", err)
	}
	if cfg2.Version != "auto2" {
		t.Errorf("version = %q, want %q", cfg2.Version, "auto2")
	}
}

func TestParseDataWithOrWithoutFormat_Whitespace(t *testing.T) {
	l := NewLoader()
	cfg, err := l.ParseDataWithOrWithoutFormat([]byte(`
		{"version": "trimmed"}
	`), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "trimmed" {
		t.Errorf("version = %q, want %q", cfg.Version, "trimmed")
	}
}

func TestParseDataWithOrWithoutFormat_InvalidFormat(t *testing.T) {
	l := NewLoader()
	_, err := l.ParseDataWithOrWithoutFormat([]byte(`{}`), "xml")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseDataWithOrWithoutFormat_Empty(t *testing.T) {
	l := NewLoader()
	_, err := l.ParseDataWithOrWithoutFormat([]byte{}, "")
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
}

func TestParseDataWithOrWithoutFormat_WhitespaceOnly(t *testing.T) {
	l := NewLoader()
	_, err := l.ParseDataWithOrWithoutFormat([]byte("   \n\t  "), "")
	if err == nil {
		t.Fatal("expected error for whitespace-only data, got nil")
	}
}

func TestParseDataWithOrWithoutFormat_InvalidJSON(t *testing.T) {
	l := NewLoader()
	_, err := l.ParseDataWithOrWithoutFormat([]byte(`{bad}`), "json")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseDataWithOrWithoutFormat_InvalidYAML(t *testing.T) {
	l := NewLoader()
	_, err := l.ParseDataWithOrWithoutFormat([]byte(`: : :`), "yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadFromInput_Stdin(t *testing.T) {
	dir := t.TempDir()
	// Simulate stdin by writing to a temp file and reading it directly
	// (we test ParseDataWithOrWithoutFormat above, so this is integration)
	l := NewLoader()

	// LoadFromInput with !isSTDIN delegates to LoadFromFile
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte(`{"version":"from-input"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := l.LoadFromInput([]string{path}, false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != "from-input" {
		t.Errorf("version = %q, want %q", cfg.Version, "from-input")
	}
}

func TestLoadFromInput_NoArgs(t *testing.T) {
	l := NewLoader()
	_, err := l.LoadFromInput(nil, false, "")
	if err == nil {
		t.Fatal("expected error for no args, got nil")
	}
}
