package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"config-analyzer/internal/config"
	"config-analyzer/internal/model"

	pb "config-analyzer/internal/server/grpc/proto"
)

// ── Fake analyzer for testing ─────────────────────────────────────────

type fakeAnalyzer struct {
	issues []model.Issue
}

func (f *fakeAnalyzer) Analyze(cfg *model.Config) []model.Issue {
	return f.issues
}

// ── Fake config loader for testing ────────────────────────────────────

type fakeLoader struct {
	cfg *model.Config
	err error
}

func (f *fakeLoader) LoadFromFile(fileName string) (*model.Config, error) {
	return f.cfg, f.err
}

func (f *fakeLoader) ParseDataWithOrWithoutFormat(data []byte, format string) (*model.Config, error) {
	return f.cfg, f.err
}

func (f *fakeLoader) LoadFromInput(args []string, isSTDIN bool, format string) (*model.Config, error) {
	return f.cfg, f.err
}

// ── Service.Run / RunE ────────────────────────────────────────────────

func TestService_Run_RequiresPath(t *testing.T) {
	svc := New(&fakeAnalyzer{}, &fakeLoader{}, 1)
	_, _, err := svc.Run(false, []string{}, "")
	if err == nil {
		t.Fatal("expected error when no path provided, got nil")
	}
}

// ── Service.Run — success paths ───────────────────────────────────────

func TestService_Run_Stdin(t *testing.T) {
	loader := &fakeLoader{
		cfg: &model.Config{
			Server: model.Server{Host: "127.0.0.1", Port: 8080},
			TLS:    model.TLS{Enabled: true},
		},
	}
	analyzer := &fakeAnalyzer{issues: []model.Issue{}}
	svc := New(analyzer, loader, 2)

	reports, summary, err := svc.Run(true, nil, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].File != "stdin" {
		t.Errorf("report.File = %q, want %q", reports[0].File, "stdin")
	}
	if summary.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", summary.TotalIssues)
	}
}

func TestService_Run_StdinWithIssues(t *testing.T) {
	loader := &fakeLoader{
		cfg: &model.Config{
			TLS: model.TLS{Enabled: false},
		},
	}
	issues := []model.Issue{
		{Severity: model.HIGH, Message: "TLS disabled", Path: "tls.enabled"},
	}
	analyzer := &fakeAnalyzer{issues: issues}
	svc := New(analyzer, loader, 2)

	reports, summary, err := svc.Run(true, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.FilesWithIssues != 1 {
		t.Errorf("FilesWithIssues = %d, want 1", summary.FilesWithIssues)
	}
	if summary.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", summary.TotalIssues)
	}
	_ = reports // used implicitly via summary
}

func TestService_Run_StdinLoaderError(t *testing.T) {
	expectedErr := fmt.Errorf("failed to parse stdin")
	loader := &fakeLoader{err: expectedErr}
	svc := New(&fakeAnalyzer{}, loader, 2)

	_, _, err := svc.Run(true, nil, "json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestService_Run_PathSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"server":{"port":8080}}`), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := &fakeAnalyzer{issues: []model.Issue{}}
	svc := New(analyzer, newRealLoader(), 2)

	_, summary, err := svc.Run(false, []string{configPath}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", summary.TotalIssues)
	}
}

// ── Service.AnalyzePath ──────────────────────────────────────────────

func TestService_AnalyzePath_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	configData := []byte(`{"server":{"host":"0.0.0.0","port":8080},"tls":{"enabled":false}}`)
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatal(err)
	}

	loader := newRealLoader()
	analyzer := &fakeAnalyzer{issues: []model.Issue{
		{Severity: model.HIGH, Message: "TLS disabled", Path: "tls.enabled"},
	}}
	svc := New(analyzer, loader, 2)

	reports, summary, err := svc.AnalyzePath(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].File != configPath {
		t.Errorf("report.File = %q, want %q", reports[0].File, configPath)
	}
	if summary.TotalIssues < 1 {
		t.Errorf("expected at least 1 issue, got %d", summary.TotalIssues)
	}
}

func TestService_AnalyzePath_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple config files
	writeFile(filepath.Join(tmpDir, "a.json"), []byte(`{"tls":{"enabled":false}}`), 0644)
	writeFile(filepath.Join(tmpDir, "b.yaml"), []byte("server:\n  host: 0.0.0.0"), 0644)
	writeFile(filepath.Join(tmpDir, "c.txt"), []byte(`not a config`), 0644) // should be skipped

	analyzer := &fakeAnalyzer{issues: []model.Issue{
		{Severity: model.HIGH, Message: "test issue", Path: "test"},
	}}
	svc := New(analyzer, newRealLoader(), 2)

	reports, summary, err := svc.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only .json and .yaml files should be analyzed
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if summary.FilesWithIssues != 2 {
		t.Errorf("FilesWithIssues = %d, want 2", summary.FilesWithIssues)
	}
}

func TestService_AnalyzePath_NonExistent(t *testing.T) {
	svc := New(&fakeAnalyzer{}, newRealLoader(), 2)
	_, _, err := svc.AnalyzePath("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestService_AnalyzePath_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	svc := New(&fakeAnalyzer{}, newRealLoader(), 2)

	reports, summary, err := svc.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reports != nil {
		t.Errorf("expected nil reports for empty dir, got %v", reports)
	}
	if summary != nil {
		t.Errorf("expected nil summary for empty dir, got %v", summary)
	}
}

// ── Service.AnalyzeStdin ─────────────────────────────────────────────

func TestService_AnalyzeStdin_NoIssues(t *testing.T) {
	loader := &fakeLoader{
		cfg: &model.Config{
			Server: model.Server{Host: "127.0.0.1", Port: 443},
			TLS:    model.TLS{Enabled: true},
			Log:    model.Log{Level: "info"},
		},
	}
	analyzer := &fakeAnalyzer{issues: []model.Issue{}}
	svc := New(analyzer, loader, 2)

	reports, summary, err := svc.AnalyzeStdin("json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if summary.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", summary.TotalIssues)
	}
	if summary.FilesWithIssues != 0 {
		t.Errorf("FilesWithIssues = %d, want 0", summary.FilesWithIssues)
	}
}

func TestService_AnalyzeStdin_MultipleIssues(t *testing.T) {
	loader := &fakeLoader{
		cfg: &model.Config{
			TLS:     model.TLS{Enabled: false},
			Server:  model.Server{Host: "0.0.0.0"},
			Log:     model.Log{Level: "debug"},
			Storage: model.Storage{DigestAlgorithm: "md5"},
		},
	}
	issues := []model.Issue{
		{Severity: model.HIGH, Message: "TLS disabled", Path: "tls.enabled"},
		{Severity: model.MEDIUM, Message: "Insecure binding", Path: "server.host"},
		{Severity: model.LOW, Message: "Debug logging", Path: "log.level"},
	}
	analyzer := &fakeAnalyzer{issues: issues}
	svc := New(analyzer, loader, 2)

	reports, summary, err := svc.AnalyzeStdin("yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", summary.TotalIssues)
	}
	if summary.FilesWithIssues != 1 {
		t.Errorf("FilesWithIssues = %d, want 1", summary.FilesWithIssues)
	}
	_ = reports
}

func TestService_AnalyzeStdin_LoaderError(t *testing.T) {
	expectedErr := fmt.Errorf("invalid format")
	loader := &fakeLoader{err: expectedErr}
	svc := New(&fakeAnalyzer{}, loader, 2)

	_, _, err := svc.AnalyzeStdin("invalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── ProtoToConfig ─────────────────────────────────────────────────────

func TestProtoToConfig_Nil(t *testing.T) {
	_, err := ProtoToConfig(nil)
	if err == nil {
		t.Fatal("expected error for nil proto config, got nil")
	}
}

func TestProtoToConfig_AllNil(t *testing.T) {
	pbCfg := &pb.Config{}

	cfg, err := ProtoToConfig(pbCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "" {
		t.Errorf("server.host = %q, want empty", cfg.Server.Host)
	}
	if cfg.Server.Port != 0 {
		t.Errorf("server.port = %d, want 0", cfg.Server.Port)
	}
	if cfg.Auth.Password != "" {
		t.Errorf("auth.password = %q, want empty", cfg.Auth.Password)
	}
	if cfg.Log.Level != "" {
		t.Errorf("log.level = %q, want empty", cfg.Log.Level)
	}
	if cfg.TLS.Enabled {
		t.Error("tls.enabled = true, want false")
	}
	if cfg.Storage.DigestAlgorithm != "" {
		t.Errorf("storage.digest-algorithm = %q, want empty", cfg.Storage.DigestAlgorithm)
	}
}

func TestProtoToConfig_OnlyTls(t *testing.T) {
	pbCfg := &pb.Config{
		Tls: &pb.Tls{Enabled: false},
	}

	cfg, err := ProtoToConfig(pbCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.TLS.Enabled {
		t.Error("tls.enabled = true, want false")
	}
}

func TestProtoToConfig_Partial(t *testing.T) {
	pbCfg := &pb.Config{
		Server: &pb.Server{Host: "127.0.0.1"},
	}

	cfg, err := ProtoToConfig(pbCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	// Other fields should be zero values
	if cfg.Auth.Password != "" {
		t.Errorf("auth.password = %q, want empty", cfg.Auth.Password)
	}
}

func TestProtoToConfig_FullConfig(t *testing.T) {
	pbCfg := &pb.Config{
		Server:  &pb.Server{Host: "localhost", Port: 8443},
		Auth:    &pb.Auth{Password: "supersecret123"},
		Log:     &pb.Log{Output: "/var/log/app.log", Level: "info"},
		Tls:     &pb.Tls{Enabled: true},
		Storage: &pb.Storage{DigestAlgorithm: "sha512"},
	}

	cfg, err := ProtoToConfig(pbCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "localhost")
	}
	if cfg.Server.Port != 8443 {
		t.Errorf("server.port = %d, want 8443", cfg.Server.Port)
	}
	if cfg.Auth.Password != "supersecret123" {
		t.Errorf("auth.password = %q, want %q", cfg.Auth.Password, "supersecret123")
	}
	if cfg.Log.Output != "/var/log/app.log" {
		t.Errorf("log.output = %q, want %q", cfg.Log.Output, "/var/log/app.log")
	}
	if cfg.Log.Level != "info" {
		t.Errorf("log.level = %q, want %q", cfg.Log.Level, "info")
	}
	if !cfg.TLS.Enabled {
		t.Error("tls.enabled = false, want true")
	}
	if cfg.Storage.DigestAlgorithm != "sha512" {
		t.Errorf("storage.digest-algorithm = %q, want %q", cfg.Storage.DigestAlgorithm, "sha512")
	}
}

func TestProtoToConfig_FullMapping(t *testing.T) {
	pbCfg := &pb.Config{
		Server:  &pb.Server{Host: "10.0.0.1", Port: 9090},
		Auth:    &pb.Auth{Password: "secret"},
		Log:     &pb.Log{Output: "stderr", Level: "warn"},
		Tls:     &pb.Tls{Enabled: true},
		Storage: &pb.Storage{DigestAlgorithm: "sha256"},
	}

	cfg, err := ProtoToConfig(pbCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Host != "10.0.0.1" {
		t.Errorf("server.host = %q, want %q", cfg.Server.Host, "10.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("server.port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Auth.Password != "secret" {
		t.Errorf("auth.password = %q, want %q", cfg.Auth.Password, "secret")
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("log.level = %q, want %q", cfg.Log.Level, "warn")
	}
	if !cfg.TLS.Enabled {
		t.Error("tls.enabled = false, want true")
	}
	if cfg.Storage.DigestAlgorithm != "sha256" {
		t.Errorf("storage.digest-algorithm = %q, want %q", cfg.Storage.DigestAlgorithm, "sha256")
	}
}

// ── Parallel analysis integration ────────────────────────────────────

func TestService_AnalyzeParallel_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 5 config files
	for i := 0; i < 5; i++ {
		content := []byte(fmt.Sprintf(`{"server":{"port":%d}}`, 8080+i))
		writeFile(filepath.Join(tmpDir, fmt.Sprintf("config%d.json", i)), content, 0644)
	}

	analyzer := &fakeAnalyzer{
		issues: []model.Issue{{Severity: model.LOW, Message: "test", Path: "test"}},
	}
	svc := New(analyzer, newRealLoader(), 3)

	reports, summary, err := svc.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 5 {
		t.Fatalf("expected 5 reports, got %d", len(reports))
	}
	if summary.TotalIssues != 5 {
		t.Errorf("TotalIssues = %d, want 5", summary.TotalIssues)
	}
}

func TestService_AnalyzeParallel_LoadError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid and an invalid file
	writeFile(filepath.Join(tmpDir, "valid.json"), []byte(`{"server":{"port":8080}}`), 0644)
	writeFile(filepath.Join(tmpDir, "invalid.json"), []byte(`{invalid json}`), 0644)

	analyzer := &fakeAnalyzer{issues: []model.Issue{}}
	svc := New(analyzer, newRealLoader(), 2)

	reports, _, err := svc.AnalyzePath(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}

	// Find the invalid file report
	var foundInvalid bool
	for _, r := range reports {
		if filepath.Base(r.File) == "invalid.json" && r.Err != nil {
			foundInvalid = true
		}
	}
	if !foundInvalid {
		t.Error("expected error report for invalid.json")
	}
}

// ── Real loader for integration-style tests ──────────────────────────

type realLoader struct {
	*config.Loader
}

func newRealLoader() *realLoader {
	return &realLoader{Loader: config.NewLoader()}
}
