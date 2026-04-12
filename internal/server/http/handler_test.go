package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"config-analyzer/internal/model"
)

// ── Fake analyzer ─────────────────────────────────────────────────────

type fakeAnalyzer struct {
	issues []model.Issue
}

func (f *fakeAnalyzer) Analyze(cfg *model.Config) []model.Issue {
	return f.issues
}

// ── Tests ─────────────────────────────────────────────────────────────

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := NewHandler(&fakeAnalyzer{})
	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	h := NewHandler(&fakeAnalyzer{})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != StatusError {
		t.Errorf("status = %q, want %q", resp.Status, StatusError)
	}
	if resp.ErrorMessage == "" {
		t.Error("expected non-empty error_message")
	}
}

func TestHandler_CleanConfig(t *testing.T) {
	h := NewHandler(&fakeAnalyzer{issues: nil})
	body := `{"server":{"host":"127.0.0.1","port":8080},"tls":{"enabled":true}}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != StatusClean {
		t.Errorf("status = %q, want %q", resp.Status, StatusClean)
	}
	if resp.Issues != nil {
		t.Errorf("expected nil issues (omitempty), got %v", resp.Issues)
	}
}

func TestHandler_ConfigWithIssues(t *testing.T) {
	issues := []model.Issue{
		{Severity: model.HIGH, Message: "TLS disabled", Path: "tls.enabled"},
		{Severity: model.MEDIUM, Message: "0.0.0.0", Path: "server.host"},
	}
	h := NewHandler(&fakeAnalyzer{issues: issues})

	body := `{"server":{"host":"0.0.0.0"},"tls":{"enabled":false}}`
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != StatusIssued {
		t.Errorf("status = %q, want %q", resp.Status, StatusIssued)
	}
	if len(resp.Issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(resp.Issues))
	}
	if resp.Issues[0].Severity != model.HIGH {
		t.Errorf("severity[0] = %q, want %q", resp.Issues[0].Severity, model.HIGH)
	}
}

func TestHandler_EmptyBody(t *testing.T) {
	h := NewHandler(&fakeAnalyzer{})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp AnalyzeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Status != StatusError {
		t.Errorf("status = %q, want %q", resp.Status, StatusError)
	}
	if resp.ErrorMessage == "" {
		t.Error("expected non-empty error_message")
	}
}

func TestRouter_RouteExists(t *testing.T) {
	h := NewHandler(&fakeAnalyzer{})
	router := NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not be 404
	if w.Code == http.StatusNotFound {
		t.Error("/analyze route not found")
	}
}
