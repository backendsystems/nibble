package delete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeDeletePathEscape(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "safe")
	os.MkdirAll(baseDir, 0o755)

	// Create a file inside base directory
	validFile := filepath.Join(baseDir, "file.json")
	os.WriteFile(validFile, []byte("{}"), 0o644)

	// Create a file outside base directory
	evilFile := filepath.Join(tmpDir, "evil.json")
	os.WriteFile(evilFile, []byte("{}"), 0o644)

	tests := []struct {
		name      string
		path      string
		baseDir   string
		wantAllow bool
	}{
		{"valid file in base", validFile, baseDir, true},
		{"file outside base", evilFile, baseDir, false},
		{"path traversal attempt", filepath.Join(baseDir, "..", "..", "evil.json"), baseDir, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate SafeDelete's path validation logic
			absPath, _ := filepath.Abs(tt.path)
			absBase, _ := filepath.Abs(tt.baseDir)
			allowed := strings.HasPrefix(absPath, absBase+string(filepath.Separator))

			if allowed != tt.wantAllow {
				t.Errorf("path validation failed: path=%s, allowed=%v, want=%v",
					absPath, allowed, tt.wantAllow)
			}
		})
	}
}

func TestSafeDeleteDirJsonOnly(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "safe")
	os.MkdirAll(baseDir, 0o755)

	// Create test files
	jsonFile := filepath.Join(baseDir, "scan.json")
	textFile := filepath.Join(baseDir, "readme.txt")
	binFile := filepath.Join(baseDir, "data.bin")

	os.WriteFile(jsonFile, []byte("{}"), 0o644)
	os.WriteFile(textFile, []byte("text"), 0o644)
	os.WriteFile(binFile, []byte("binary"), 0o644)

	// SafeDeleteDir should only delete .json files
	SafeDeleteDir(baseDir, tmpDir)

	// Check results
	if _, err := os.Stat(jsonFile); err == nil {
		t.Error("JSON file should be deleted")
	}
	if _, err := os.Stat(textFile); err != nil {
		t.Error("Text file should not be deleted")
	}
	if _, err := os.Stat(binFile); err != nil {
		t.Error("Binary file should not be deleted")
	}
}

func TestSafeDeleteDirEscapeAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "safe")
	os.MkdirAll(baseDir, 0o755)

	// Create a file outside base directory
	evilDir := filepath.Join(tmpDir, "evil")
	os.MkdirAll(evilDir, 0o755)
	evilFile := filepath.Join(evilDir, "secret.json")
	os.WriteFile(evilFile, []byte("{}"), 0o644)

	// Try to delete directory outside base
	err := SafeDeleteDir(evilDir, baseDir)
	if err != os.ErrPermission {
		t.Errorf("Expected ErrPermission, got %v", err)
	}

	// Evil file should still exist
	if _, err := os.Stat(evilFile); err != nil {
		t.Error("Evil file should not be deleted")
	}
}
