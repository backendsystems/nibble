package parameters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAndNormalizeCIDR(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"192.168.1.1", "192.168.1.1/32", false},
		{"10.0.0.0/24", "10.0.0.0/24", false},
		{"172.16.0.0/16", "172.16.0.0/16", false},
		{"192.168.1.0/32", "192.168.1.0/32", false},
		{"not-an-ip", "", true},
		{"::1", "", true},
		{"192.168.1.0/33", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := validateAndNormalizeCIDR(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndNormalizeCIDR(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateAndNormalizeCIDR(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadTargetsFile(t *testing.T) {
	// Test with the bundled example.txt
	targets, err := readTargetsFile("example.txt")
	if err != nil {
		t.Fatalf("readTargetsFile(example.txt) error: %v", err)
	}

	want := []string{
		"192.168.1.1/32",
		"192.168.1.0/24",
		"10.0.0.0/16",
		"172.16.0.100/32",
	}

	if len(targets) != len(want) {
		t.Fatalf("got %d targets, want %d", len(targets), len(want))
	}
	for i := range want {
		if targets[i] != want[i] {
			t.Errorf("target[%d] = %q, want %q", i, targets[i], want[i])
		}
	}
}

func TestReadTargetsFileEmpty(t *testing.T) {
	f := filepath.Join(t.TempDir(), "empty.txt")
	os.WriteFile(f, []byte("# only comments\n\n"), 0644)

	_, err := readTargetsFile(f)
	if err == nil {
		t.Error("expected error for empty targets file")
	}
}

func TestReadTargetsFileInvalidLine(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.txt")
	os.WriteFile(f, []byte("192.168.1.1\nnot-valid\n"), 0644)

	_, err := readTargetsFile(f)
	if err == nil {
		t.Error("expected error for invalid line in targets file")
	}
}

func TestResolveTargetsDirect(t *testing.T) {
	targets, err := resolveTargets("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 || targets[0] != "192.168.1.0/24" {
		t.Errorf("got %v, want [192.168.1.0/24]", targets)
	}
}

func TestResolveTargetsFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "targets.txt")
	os.WriteFile(f, []byte("10.0.0.1\n10.0.0.2\n"), 0644)

	targets, err := resolveTargets(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("got %d targets, want 2", len(targets))
	}
	if targets[0] != "10.0.0.1/32" || targets[1] != "10.0.0.2/32" {
		t.Errorf("got %v", targets)
	}
}
