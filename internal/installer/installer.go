package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Installer defines the interface for installing Go modules.
type Installer interface {
	Install(modulePath string, version string) (string, error)
}

// DefaultInstaller implements Installer using go install.
type DefaultInstaller struct {
	goproxy    string
	cgoenabled *bool
}

// NewDefaultInstaller creates a new DefaultInstaller.
func NewDefaultInstaller() *DefaultInstaller {
	return &DefaultInstaller{}
}

// NewDefaultInstallerWithGOPROXY creates a new DefaultInstaller that overrides the
// GOPROXY environment variable with the provided value when it is non-empty.
func NewDefaultInstallerWithGOPROXY(goproxy string) *DefaultInstaller {
	return &DefaultInstaller{goproxy: goproxy}
}

// NewDefaultInstallerWithOptions creates a new DefaultInstaller with the provided
// installer options. Non-empty goproxy overrides the GOPROXY environment variable;
// non-nil cgoenabled overrides the CGO_ENABLED environment variable.
func NewDefaultInstallerWithOptions(goproxy string, cgoenabled *bool) *DefaultInstaller {
	return &DefaultInstaller{goproxy: goproxy, cgoenabled: cgoenabled}
}

// buildInstallCmd creates the exec.Cmd for "go install {modulePath}@{version}" with the
// current process environment so that variables such as GOPROXY are forwarded.
// If the installer was configured with a GOPROXY value it overrides any inherited GOPROXY.
// If the installer was configured with a CGO_ENABLED value it overrides any inherited CGO_ENABLED.
func (d *DefaultInstaller) buildInstallCmd(modulePath, version string) *exec.Cmd {
	cmd := exec.Command("go", "install", modulePath+"@"+version)
	env := os.Environ()
	if d.goproxy != "" {
		filtered := make([]string, 0, len(env))
		for _, e := range env {
			if !strings.HasPrefix(e, "GOPROXY=") {
				filtered = append(filtered, e)
			}
		}
		env = append(filtered, "GOPROXY="+d.goproxy)
	}
	if d.cgoenabled != nil {
		value := "1"
		if !*d.cgoenabled {
			value = "0"
		}
		filtered := make([]string, 0, len(env))
		for _, e := range env {
			if !strings.HasPrefix(e, "CGO_ENABLED=") {
				filtered = append(filtered, e)
			}
		}
		env = append(filtered, "CGO_ENABLED="+value)
	}
	cmd.Env = env
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
