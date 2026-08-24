package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRemoveArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantName   string
		wantDelete bool
		wantErr    bool
	}{
		{name: "name only", args: []string{"tool"}, wantName: "tool"},
		{name: "delete after name is rejected", args: []string{"tool", "--delete"}, wantErr: true},
		{name: "delete before name", args: []string{"--delete", "tool"}, wantName: "tool", wantDelete: true},
		{name: "missing name", args: nil, wantErr: true},
		{name: "unknown flag", args: []string{"--force", "tool"}, wantErr: true},
		{name: "multiple names", args: []string{"tool", "other"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRemoveArgs(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.name != tc.wantName || got.deleteBinary != tc.wantDelete {
				t.Fatalf("parseRemoveArgs(%q) = %+v, want name=%q delete=%t", tc.args, got, tc.wantName, tc.wantDelete)
			}
		})
	}
}

func TestPrintRemoveHelp(t *testing.T) {
	var output bytes.Buffer
	printRemoveHelp(&output)
	if !strings.Contains(output.String(), "gogitup remove [--delete]") || !strings.Contains(output.String(), "--delete") {
		t.Fatalf("unexpected help output %q", output.String())
	}
}

func TestDeleteInstalledBinary(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "tool")
	if err := os.WriteFile(binaryPath, []byte("binary"), 0700); err != nil {
		t.Fatalf("write test binary: %v", err)
	}
	t.Setenv("PATH", dir)

	if err := deleteInstalledBinary("tool"); err != nil {
		t.Fatalf("deleteInstalledBinary() error = %v", err)
	}
	if _, err := os.Stat(binaryPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("binary still exists or could not be checked: %v", err)
	}
}

func TestDeleteInstalledBinaryNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := deleteInstalledBinary("missing-tool"); err == nil {
		t.Fatal("expected an error")
	}
}
