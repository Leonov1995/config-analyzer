package analyzer

import (
	"testing"

	"config-analyzer/internal/model"
	"config-analyzer/internal/rule"
)

func TestAnalyze_NilConfig(t *testing.T) {
	a := New([]rule.Rule{})
	issues := a.Analyze(nil)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue for nil config, got %d", len(issues))
	}
	if issues[0].Severity != model.HIGH {
		t.Errorf("severity = %q, want %q", issues[0].Severity, model.HIGH)
	}
	if issues[0].Message != "nil config" {
		t.Errorf("message = %q, want %q", issues[0].Message, "nil config")
	}
}

func TestAnalyze_CleanConfig(t *testing.T) {
	rules := rule.NewDefaultRules()
	a := New(rules)

	cfg := &model.Config{
		Version: "1.0",
		Server:  model.Server{Host: "127.0.0.1", Port: 8080},
		TLS:     model.TLS{Enabled: true},
		Auth:    model.Auth{Password: "strongpassword123"},
		Log:     model.Log{Level: "info", Output: "stdout"},
		Storage: model.Storage{DigestAlgorithm: "sha256"},
	}

	issues := a.Analyze(cfg)
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for clean config, got %d: %v", len(issues), issues)
	}
}

func TestAnalyze_AllIssues(t *testing.T) {
	rules := rule.NewDefaultRules()
	a := New(rules)

	cfg := &model.Config{
		Version: "1.0",
		Server:  model.Server{Host: "0.0.0.0", Port: 8080},
		TLS:     model.TLS{Enabled: false},
		Auth:    model.Auth{Password: "123"},
		Log:     model.Log{Level: "debug", Output: "stdout"},
		Storage: model.Storage{DigestAlgorithm: "md5"},
	}

	issues := a.Analyze(cfg)

	// 5 rules should all trigger
	if len(issues) != 5 {
		t.Errorf("expected 5 issues, got %d", len(issues))
	}

	severitySet := map[model.Severity]int{}
	for _, i := range issues {
		severitySet[i.Severity]++
	}

	if severitySet[model.HIGH] != 3 {
		t.Errorf("expected 3 HIGH, got %d", severitySet[model.HIGH])
	}
	if severitySet[model.MEDIUM] != 1 {
		t.Errorf("expected 1 MEDIUM, got %d", severitySet[model.MEDIUM])
	}
	if severitySet[model.LOW] != 1 {
		t.Errorf("expected 1 LOW, got %d", severitySet[model.LOW])
	}
}

func TestAnalyze_EmptyConfig(t *testing.T) {
	rules := rule.NewDefaultRules()
	a := New(rules)

	cfg := &model.Config{}
	issues := a.Analyze(cfg)

	// Empty config should trigger: TLS (false default), empty host, empty password (not weak since empty), debug (empty != debug), empty storage
	// Actually: TLS=false→HIGH, host=""→MEDIUM, password=""→no issue (empty skips), log=""→no issue (empty != debug), storage=""→no issue
	if len(issues) < 2 {
		t.Errorf("expected at least 2 issues for empty config, got %d", len(issues))
	}
}

func TestAnalyze_CustomRules(t *testing.T) {
	customRule := &testRule{trigger: true}
	a := New([]rule.Rule{customRule})

	cfg := &model.Config{}
	issues := a.Analyze(cfg)

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue from custom rule, got %d", len(issues))
	}
	if issues[0].Message != "custom rule triggered" {
		t.Errorf("message = %q, want %q", issues[0].Message, "custom rule triggered")
	}

	// Custom rule that doesn't trigger
	customRule2 := &testRule{trigger: false}
	a2 := New([]rule.Rule{customRule2})
	issues2 := a2.Analyze(cfg)
	if len(issues2) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues2))
	}
}

type testRule struct{ trigger bool }

func (r *testRule) Check(cfg *model.Config) *model.Issue {
	if r.trigger {
		return &model.Issue{
			Severity: model.HIGH,
			Message:  "custom rule triggered",
		}
	}
	return nil
}
