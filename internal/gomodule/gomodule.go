package gomodule

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Result describes the Go toolchain's update decision for an installed module.
type Result struct {
	LatestVersion   string
	UpdateAvailable bool
}

// Resolver checks module updates through the Go toolchain.
type Resolver interface {
	Check(modulePath, installedVersion string) (Result, error)
}

// DefaultResolver implements Resolver using go list.
type DefaultResolver struct {
	goproxy string
}

// NewDefaultResolver creates a resolver that inherits GOPROXY from the environment.
func NewDefaultResolver() *DefaultResolver {
	return &DefaultResolver{}
}

// NewDefaultResolverWithGOPROXY creates a resolver that overrides GOPROXY when
// the provided value is non-empty.
func NewDefaultResolverWithGOPROXY(goproxy string) *DefaultResolver {
	return &DefaultResolver{goproxy: goproxy}
}

type moduleInfo struct {
	Version string      `json:"Version"`
	Update  *moduleInfo `json:"Update"`
}

// Check asks the Go toolchain whether a newer module version is available.
func (r *DefaultResolver) Check(modulePath, installedVersion string) (Result, error) {
	cmd := r.buildListCmd(modulePath, installedVersion)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Result{}, fmt.Errorf("go list -m -u %s@%s failed: %w\n%s", modulePath, installedVersion, err, string(out))
	}

	return ParseUpdate(out)
}

func (r *DefaultResolver) buildListCmd(modulePath, installedVersion string) *exec.Cmd {
	cmd := exec.Command("go", "list", "-m", "-u", "-json", modulePath+"@"+installedVersion)
	env := os.Environ()
	if r.goproxy != "" {
		filtered := make([]string, 0, len(env))
		for _, entry := range env {
			if !strings.HasPrefix(entry, "GOPROXY=") {
				filtered = append(filtered, entry)
			}
		}
		env = append(filtered, "GOPROXY="+r.goproxy)
	}
	cmd.Env = env
	return cmd
}

// ParseUpdate extracts the Go toolchain's update decision from go list JSON.
func ParseUpdate(data []byte) (Result, error) {
	var info moduleInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return Result{}, fmt.Errorf("failed to parse go list output: %w", err)
	}
	if info.Version == "" {
		return Result{}, errors.New("go list output did not include a version")
	}
	if info.Update == nil {
		return Result{LatestVersion: info.Version}, nil
	}
	if info.Update.Version == "" {
		return Result{}, errors.New("go list update output did not include a version")
	}
	return Result{
		LatestVersion:   info.Update.Version,
		UpdateAvailable: true,
	}, nil
}
