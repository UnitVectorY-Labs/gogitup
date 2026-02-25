package installer

import (
	"fmt"
	"testing"
)

// mockInstaller is a test double for the Installer interface.
type mockInstaller struct {
	output string
	err    error
}

func (m *mockInstaller) Install(modulePath string, version string) (string, error) {
	return m.output, m.err
}

// TestMockImplementsInterface verifies that mockInstaller satisfies the Installer interface.
func TestMockImplementsInterface(t *testing.T) {
	var _ Installer = (*mockInstaller)(nil)
}

// TestDefaultInstallerImplementsInterface verifies that DefaultInstaller satisfies the Installer interface.
func TestDefaultInstallerImplementsInterface(t *testing.T) {
	var _ Installer = (*DefaultInstaller)(nil)
}

// TestNewDefaultInstaller verifies that the constructor returns a non-nil instance.
func TestNewDefaultInstaller(t *testing.T) {
	inst := NewDefaultInstaller()
	if inst == nil {
		t.Fatalf("expected non-nil DefaultInstaller")
	}
}

// TestMockInstallerSuccess verifies a successful install via the mock.
func TestMockInstallerSuccess(t *testing.T) {
	m := &mockInstaller{output: "installed successfully", err: nil}
	out, err := m.Install("github.com/example/tool", "v1.0.0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "installed successfully" {
		t.Fatalf("expected 'installed successfully', got %q", out)
	}
}

// TestMockInstallerFailure verifies error handling via the mock.
func TestMockInstallerFailure(t *testing.T) {
	m := &mockInstaller{output: "module not found", err: fmt.Errorf("exit status 1")}
	out, err := m.Install("github.com/example/tool", "v0.0.0")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if out != "module not found" {
		t.Fatalf("expected 'module not found', got %q", out)
	}
}
