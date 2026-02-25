package goversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Info holds version information about an installed Go binary.
type Info struct {
	Path      string
	Version   string
	GoVersion string
}

// Runner is an interface for retrieving version info from Go binaries.
type Runner interface {
	GetInfo(binaryName string) (*Info, error)
}

// DefaultRunner implements Runner by executing the go version command.
type DefaultRunner struct{}

// versionOutput represents a single entry in the go version -m -json output.
type versionOutput struct {
	Path string `json:"Path"`
	Main struct {
		Path    string `json:"Path"`
		Version string `json:"Version"`
	} `json:"Main"`
	GoVersion string `json:"GoVersion"`
}

// GetInfo runs go version -m -json against the named binary and returns its Info.
func (d *DefaultRunner) GetInfo(binaryName string) (*Info, error) {
	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		return nil, errors.New("binary not found: " + binaryName)
	}

	cmd := exec.Command("go", "version", "-m", "-json", binaryPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New("failed to execute go version")
	}

	return ParseVersionJSON(output)
}

// ParseVersionJSON parses the JSON output from go version -m -json into Info.
func ParseVersionJSON(data []byte) (*Info, error) {
	var entries []versionOutput
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse go version output: %w", err)
	}

	if len(entries) == 0 {
		return nil, errors.New("binary was not installed with go install")
	}

	entry := entries[0]
	if entry.Main.Path == "" || entry.Main.Version == "" {
		return nil, errors.New("binary was not installed with go install")
	}

	return &Info{
		Path:      entry.Main.Path,
		Version:   entry.Main.Version,
		GoVersion: entry.GoVersion,
	}, nil
}

// IsGitHubRepo checks if the module path starts with "github.com/".
func IsGitHubRepo(modulePath string) bool {
	return strings.HasPrefix(modulePath, "github.com/")
}

// ParseGitHubRepo extracts the owner and repo from a "github.com/owner/repo" path.
func ParseGitHubRepo(modulePath string) (owner string, repo string, err error) {
	if !IsGitHubRepo(modulePath) {
		return "", "", errors.New("not a GitHub repository path: " + modulePath)
	}

	parts := strings.SplitN(modulePath, "/", 4)
	if len(parts) < 3 || parts[1] == "" || parts[2] == "" {
		return "", "", errors.New("invalid GitHub repository path: " + modulePath)
	}

	return parts[1], parts[2], nil
}
