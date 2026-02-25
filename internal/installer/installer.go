package installer

import (
	"fmt"
	"os"
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

// buildInstallCmd creates the exec.Cmd for "go install {modulePath}@{version}" with the
// current process environment so that variables such as GOPROXY are forwarded.
func (d *DefaultInstaller) buildInstallCmd(modulePath, version string) *exec.Cmd {
	cmd := exec.Command("go", "install", modulePath+"@"+version)
	cmd.Env = os.Environ()
	return cmd
}

// Install runs "go install {modulePath}@{version}" and returns the combined output.
func (d *DefaultInstaller) Install(modulePath string, version string) (string, error) {
	cmd := d.buildInstallCmd(modulePath, version)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("go install %s@%s failed: %w\n%s", modulePath, version, err, string(out))
	}
	return string(out), nil
}
