package service

import (
	"os"
	"path/filepath"
	"testing"

	"config-analyzer/internal/model"
)

func TestCollectFiles_SingleFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "config.json")
	if err := writeFile(path, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := collectFiles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0] != path {
		t.Errorf("path = %q, want %q", files[0], path)
	}
}

func TestCollectFiles_Directory(t *testing.T) {
	tmp := t.TempDir()

	// Create files
	writeFile(filepath.Join(tmp, "a.json"), []byte(`{}`), 0644)
	writeFile(filepath.Join(tmp, "b.yaml"), []byte(`{}`), 0644)
	writeFile(filepath.Join(tmp, "c.yml"), []byte(`{}`), 0644)
	writeFile(filepath.Join(tmp, "d.txt"), []byte(`{}`), 0644) // should be skipped

	files, err := collectFiles(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("expected 3 config files, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_NestedDirectory(t *testing.T) {
	tmp := t.TempDir()

	writeFile(filepath.Join(tmp, "root.json"), []byte(`{}`), 0644)
	subdir := filepath.Join(tmp, "sub")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(filepath.Join(subdir, "nested.yaml"), []byte(`{}`), 0644)

	files, err := collectFiles(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestCollectFiles_NonExistent(t *testing.T) {
	_, err := collectFiles("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestCollectFiles_EmptyDir(t *testing.T) {
	tmp := t.TempDir()
	files, err := collectFiles(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty dir, got %d", len(files))
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"config.json", true},
		{"config.yaml", true},
		{"config.yml", true},
		{"config.JSON", true},
		{"config.YAML", true},
		{"config.txt", false},
		{"config.xml", false},
		{"config", false},
		{"/path/to/config.yaml", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isConfigFile(tt.path)
			if got != tt.want {
				t.Errorf("isConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestBuildSummary(t *testing.T) {
	reports := []model.Report{
		{File: "a.json", Issues: []model.Issue{{}, {}}},
		{File: "b.yaml", Issues: []model.Issue{{}}},
		{File: "c.json", Issues: nil},
		{File: "d.yaml", Issues: []model.Issue{}},
		{File: "e.json", Err: nil},
	}

	summary := buildSummary(reports)

	if summary.FilesWithIssues != 2 {
		t.Errorf("FilesWithIssues = %d, want 2", summary.FilesWithIssues)
	}
	if summary.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", summary.TotalIssues)
	}
}

func TestBuildSummary_Empty(t *testing.T) {
	summary := buildSummary(nil)
	if summary.FilesWithIssues != 0 {
		t.Errorf("FilesWithIssues = %d, want 0", summary.FilesWithIssues)
	}
	if summary.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", summary.TotalIssues)
	}
}

func TestBuildSummary_AllClean(t *testing.T) {
	reports := []model.Report{
		{File: "a.json", Issues: nil},
		{File: "b.yaml", Issues: []model.Issue{}},
	}

	summary := buildSummary(reports)
	if summary.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", summary.TotalIssues)
	}
}

// writeFile is a portable helper.
func writeFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
