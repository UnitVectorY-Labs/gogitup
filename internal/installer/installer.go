package installer

import (
	"fmt"
	"os/exec"
)

// Installer defines the interface for installing Go modules.
type Installer interface {
	Install(modulePath string, version string) (string, error)
}

// DefaultInstaller implements Installer using go install.
type DefaultInstaller struct{}

// NewDefaultInstaller creates a new DefaultInstaller.
func NewDefaultInstaller() *DefaultInstaller {
	return &DefaultInstaller{}
}

// Install runs "go install {modulePath}@{version}" and returns the combined output.
func (d *DefaultInstaller) Install(modulePath string, version string) (string, error) {
	cmd := exec.Command("go", "install", modulePath+"@"+version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("go install %s@%s failed: %w\n%s", modulePath, version, err, string(out))
	}
	return string(out), nil
}
