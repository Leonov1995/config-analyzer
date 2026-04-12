package grpc

import (
	"context"
	"testing"

	"config-analyzer/internal/model"
	pb "config-analyzer/internal/server/grpc/proto"
)

// ── Fake analyzer ─────────────────────────────────────────────────────

type fakeAnalyzer struct {
	issues []model.Issue
	err    error
}

func (f *fakeAnalyzer) Analyze(cfg *model.Config) []model.Issue {
	return f.issues
}

// ── Tests ─────────────────────────────────────────────────────────────

func TestServer_Analyze_CleanConfig(t *testing.T) {
	srv := NewServer(&fakeAnalyzer{issues: nil})

	req := &pb.ConfigRequest{
		Config: &pb.Config{
			Server: &pb.Server{Host: "127.0.0.1", Port: 8080},
			Tls:    &pb.Tls{Enabled: true},
		},
	}

	resp, err := srv.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != pb.Status_clean {
		t.Errorf("status = %v, want %v", resp.Status, pb.Status_clean)
	}
	if len(resp.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(resp.Issues))
	}
}

func TestServer_Analyze_ConfigWithIssues(t *testing.T) {
	issues := []model.Issue{
		{Severity: model.HIGH, Message: "TLS off", Path: "tls.enabled"},
	}
	srv := NewServer(&fakeAnalyzer{issues: issues})

	req := &pb.ConfigRequest{
		Config: &pb.Config{
			Tls: &pb.Tls{Enabled: false},
		},
	}

	resp, err := srv.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != pb.Status_issues {
		t.Errorf("status = %v, want %v", resp.Status, pb.Status_issues)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(resp.Issues))
	}
	if resp.Issues[0].Severity != "HIGH" {
		t.Errorf("severity = %q, want %q", resp.Issues[0].Severity, "HIGH")
	}
	if resp.Issues[0].Message != "TLS off" {
		t.Errorf("message = %q, want %q", resp.Issues[0].Message, "TLS off")
	}
}

func TestServer_Analyze_NilConfig(t *testing.T) {
	srv := NewServer(&fakeAnalyzer{issues: nil})

	req := &pb.ConfigRequest{Config: nil}
	resp, err := srv.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != pb.Status_error {
		t.Errorf("status = %v, want %v", resp.Status, pb.Status_error)
	}
	if resp.ErrorMessage == "" {
		t.Error("expected non-empty error_message")
	}
}

func TestServer_Analyze_EmptyRequest(t *testing.T) {
	srv := NewServer(&fakeAnalyzer{issues: nil})

	req := &pb.ConfigRequest{} // Config field is nil
	resp, err := srv.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != pb.Status_error {
		t.Errorf("status = %v, want %v", resp.Status, pb.Status_error)
	}
	if resp.ErrorMessage == "" {
		t.Error("expected non-empty error_message")
	}
}

func TestServer_Analyze_FullConfig(t *testing.T) {
	srv := NewServer(&fakeAnalyzer{issues: nil})

	req := &pb.ConfigRequest{
		Config: &pb.Config{
			Server:  &pb.Server{Host: "10.0.0.1", Port: 9090},
			Auth:    &pb.Auth{Password: "secret"},
			Log:     &pb.Log{Output: "stderr", Level: "info"},
			Tls:     &pb.Tls{Enabled: true},
			Storage: &pb.Storage{DigestAlgorithm: "sha256"},
		},
	}

	resp, err := srv.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != pb.Status_clean {
		t.Errorf("status = %v, want %v", resp.Status, pb.Status_clean)
	}
}
