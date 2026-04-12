package rule

import (
	"os"
	"testing"

	"config-analyzer/internal/model"
)

// ── TLS Rule ──────────────────────────────────────────────────────────

func TestTLSRule_Triggered(t *testing.T) {
	cfg := &model.Config{TLS: model.TLS{Enabled: false}}
	issue := TLSRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue when TLS disabled, got nil")
	}
	if issue.Severity != model.HIGH {
		t.Errorf("severity = %q, want %q", issue.Severity, model.HIGH)
	}
	if issue.Path != "tls.enabled" {
		t.Errorf("path = %q, want %q", issue.Path, "tls.enabled")
	}
}

func TestTLSRule_NotTriggered(t *testing.T) {
	cfg := &model.Config{TLS: model.TLS{Enabled: true}}
	issue := TLSRule{}.Check(cfg)
	if issue != nil {
		t.Errorf("expected nil issue, got %+v", issue)
	}
}

// ── Log Rule ──────────────────────────────────────────────────────────

func TestLogRule_DebugTriggered(t *testing.T) {
	cfg := &model.Config{Log: model.Log{Level: "debug"}}
	issue := LogRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for debug level, got nil")
	}
	if issue.Severity != model.LOW {
		t.Errorf("severity = %q, want %q", issue.Severity, model.LOW)
	}
}

func TestLogRule_CaseInsensitive(t *testing.T) {
	cfg := &model.Config{Log: model.Log{Level: "DEBUG"}}
	issue := LogRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for DEBUG (uppercase), got nil")
	}
}

func TestLogRule_WithWhitespace(t *testing.T) {
	cfg := &model.Config{Log: model.Log{Level: "  debug  "}}
	issue := LogRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for '  debug  ', got nil")
	}
}

func TestLogRule_OtherLevels(t *testing.T) {
	for _, level := range []string{"info", "warn", "error", "trace", ""} {
		t.Run(level, func(t *testing.T) {
			cfg := &model.Config{Log: model.Log{Level: level}}
			issue := LogRule{}.Check(cfg)
			if issue != nil {
				t.Errorf("expected nil for level %q, got %+v", level, issue)
			}
		})
	}
}

// ── Storage Rule ──────────────────────────────────────────────────────

func TestStorageRule_MD5Triggered(t *testing.T) {
	cfg := &model.Config{Storage: model.Storage{DigestAlgorithm: "md5"}}
	issue := StorageRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for md5, got nil")
	}
	if issue.Severity != model.HIGH {
		t.Errorf("severity = %q, want %q", issue.Severity, model.HIGH)
	}
}

func TestStorageRule_SHA1Triggered(t *testing.T) {
	cfg := &model.Config{Storage: model.Storage{DigestAlgorithm: "sha1"}}
	issue := StorageRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for sha1, got nil")
	}
}

func TestStorageRule_CaseInsensitive(t *testing.T) {
	for _, algo := range []string{"MD5", "Sha1", "SHA1", "Md5"} {
		t.Run(algo, func(t *testing.T) {
			cfg := &model.Config{Storage: model.Storage{DigestAlgorithm: algo}}
			issue := StorageRule{}.Check(cfg)
			if issue == nil {
				t.Errorf("expected issue for %q, got nil", algo)
			}
		})
	}
}

func TestStorageRule_SafeAlgorithms(t *testing.T) {
	for _, algo := range []string{"sha256", "sha512", "blake2", ""} {
		t.Run(algo, func(t *testing.T) {
			cfg := &model.Config{Storage: model.Storage{DigestAlgorithm: algo}}
			issue := StorageRule{}.Check(cfg)
			if issue != nil {
				t.Errorf("expected nil for %q, got %+v", algo, issue)
			}
		})
	}
}

// ── Password Rule ─────────────────────────────────────────────────────

func TestPasswordRule_TooShort(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "123"}}
	issue := PasswordRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for short password, got nil")
	}
	if issue.Severity != model.HIGH {
		t.Errorf("severity = %q, want %q", issue.Severity, model.HIGH)
	}
}

func TestPasswordRule_ExactEight(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "12345678"}}
	issue := PasswordRule{}.Check(cfg)
	if issue != nil {
		t.Errorf("expected nil for 8-char password, got %+v", issue)
	}
}

func TestPasswordRule_CommonPassword(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "password"}}
	issue := PasswordRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for 'password', got nil")
	}
}

func TestPasswordRule_Empty(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: ""}}
	issue := PasswordRule{}.Check(cfg)
	if issue != nil {
		t.Errorf("expected nil for empty password, got %+v", issue)
	}
}

func TestPasswordRule_Strong(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "Str0ng!P@ssw0rd"}}
	issue := PasswordRule{}.Check(cfg)
	if issue != nil {
		t.Errorf("expected nil for strong password, got %+v", issue)
	}
}

func TestPasswordRule_WhitespaceOnly(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "   "}}
	issue := PasswordRule{}.Check(cfg)
	if issue != nil {
		t.Errorf("expected nil for whitespace-only password, got %+v", issue)
	}
}

func TestPasswordRule_ShortAfterTrim(t *testing.T) {
	cfg := &model.Config{Auth: model.Auth{Password: "  abc  "}}
	issue := PasswordRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for short password with spaces, got nil")
	}
}

// ── Server Rule ───────────────────────────────────────────────────────

func TestServerRule_EmptyHost(t *testing.T) {
	cfg := &model.Config{Server: model.Server{Host: ""}}
	issue := ServerRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for empty host, got nil")
	}
	if issue.Severity != model.MEDIUM {
		t.Errorf("severity = %q, want %q", issue.Severity, model.MEDIUM)
	}
}

func TestServerRule_WildcardHost(t *testing.T) {
	cfg := &model.Config{Server: model.Server{Host: "0.0.0.0"}}
	issue := ServerRule{}.Check(cfg)
	if issue == nil {
		t.Fatal("expected issue for 0.0.0.0, got nil")
	}
	if issue.Severity != model.MEDIUM {
		t.Errorf("severity = %q, want %q", issue.Severity, model.MEDIUM)
	}
}

func TestServerRule_SafeHost(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "localhost", "10.0.0.1", "::1"} {
		t.Run(host, func(t *testing.T) {
			cfg := &model.Config{Server: model.Server{Host: host}}
			issue := ServerRule{}.Check(cfg)
			if issue != nil {
				t.Errorf("expected nil for host %q, got %+v", host, issue)
			}
		})
	}
}

// ── Permissions Rule ──────────────────────────────────────────────────

// mockFileInfo implements fileInfo for permission tests without FS access.
type mockFileInfo struct {
	perm os.FileMode
}

func (m mockFileInfo) Mode() os.FileMode { return m.perm }

func TestCheckFile_WorldWritable(t *testing.T) {
	issue := checkFileFromInfo("/path/to/config.json", mockFileInfo{perm: 0666})
	if issue == nil {
		t.Fatal("expected issue for world-writable file, got nil")
	}
	if issue.Severity != model.HIGH {
		t.Errorf("severity = %q, want %q", issue.Severity, model.HIGH)
	}
}

func TestCheckFile_NonExistent(t *testing.T) {
	// Override stat to simulate file-not-found error
	orig := stat
	stat = func(path string) (fileInfo, error) {
		return nil, os.ErrNotExist
	}
	defer func() { stat = orig }()

	issue := CheckFile("/nonexistent/config.json")
	if issue != nil {
		t.Errorf("expected nil for non-existent file, got %+v", issue)
	}
}

func TestCheckFile_SafePermissions(t *testing.T) {
	issue := checkFileFromInfo("/path/to/config.json", mockFileInfo{perm: 0600})
	if issue != nil {
		t.Errorf("expected nil for 0600 permissions, got %+v", issue)
	}
}

