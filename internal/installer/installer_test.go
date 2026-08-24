package installer

import (
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// mockInstaller is a test double for the Installer interface.
type mockInstaller struct {
	output string
	err    error
}

func (m *mockInstaller) Install(modulePath string, version string, options ...InstallOptions) (string, error) {
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

// TestNewDefaultInstallerWithGOPROXY verifies that a configured GOPROXY value overrides the environment variable.
func TestNewDefaultInstallerWithGOPROXY(t *testing.T) {
	const configProxy = "https://config-proxy.example.com"
	t.Setenv("GOPROXY", "https://env-proxy.example.com")

	inst := NewDefaultInstallerWithGOPROXY(configProxy)
	if inst == nil {
		t.Fatalf("expected non-nil DefaultInstaller")
	}

	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	found := false
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "GOPROXY=") {
			if e == "GOPROXY="+configProxy {
				found = true
			} else {
				t.Fatalf("unexpected GOPROXY value: %s (wanted %s)", e, "GOPROXY="+configProxy)
			}
			break
		}
	}
	if !found {
		t.Fatalf("GOPROXY=%s not found in command environment", configProxy)
	}
}

// TestNewDefaultInstallerWithEmptyGOPROXY verifies that an empty GOPROXY value preserves the inherited environment variable.
func TestNewDefaultInstallerWithEmptyGOPROXY(t *testing.T) {
	const envProxy = "https://env-proxy.example.com"
	t.Setenv("GOPROXY", envProxy)

	inst := NewDefaultInstallerWithGOPROXY("")
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	found := false
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "GOPROXY=") {
			if e == "GOPROXY="+envProxy {
				found = true
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected GOPROXY=%s to be inherited from environment", envProxy)
	}
}
func TestBuildInstallCmdEnvIncludesGOPROXY(t *testing.T) {
	const proxyURL = "https://proxy.example.com"
	t.Setenv("GOPROXY", proxyURL)

	inst := NewDefaultInstaller()
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	found := false
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "GOPROXY=") {
			if e == "GOPROXY="+proxyURL {
				found = true
			} else {
				t.Fatalf("unexpected GOPROXY value: %s", e)
			}
			break
		}
	}
	if !found {
		t.Fatalf("GOPROXY=%s not found in command environment", proxyURL)
	}
}

// TestBuildInstallCmdArgs verifies the go install command arguments are constructed correctly.
func TestBuildInstallCmdArgs(t *testing.T) {
	inst := NewDefaultInstaller()
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.2.3")

	args := cmd.Args
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[1] != "install" {
		t.Fatalf("expected args[1]='install', got %q", args[1])
	}
	if args[2] != "github.com/example/tool@v1.2.3" {
		t.Fatalf("expected args[2]='github.com/example/tool@v1.2.3', got %q", args[2])
	}
}

// TestNewDefaultInstallerWithOptionsCGODisabled verifies that cgo_enabled=false sets CGO_ENABLED=0.
func TestNewDefaultInstallerWithOptionsCGODisabled(t *testing.T) {
	t.Setenv("CGO_ENABLED", "1")
	cgoDisabled := false
	inst := NewDefaultInstallerWithOptions("", &cgoDisabled)
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	found := false
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "CGO_ENABLED=") {
			if e != "CGO_ENABLED=0" {
				t.Fatalf("expected CGO_ENABLED=0, got %s", e)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("CGO_ENABLED not found in command environment")
	}
}

// TestNewDefaultInstallerWithOptionsCGOEnabled verifies that cgo_enabled=true sets CGO_ENABLED=1.
func TestNewDefaultInstallerWithOptionsCGOEnabled(t *testing.T) {
	t.Setenv("CGO_ENABLED", "0")
	cgoEnabled := true
	inst := NewDefaultInstallerWithOptions("", &cgoEnabled)
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	found := false
	for _, e := range cmd.Env {
		if strings.HasPrefix(e, "CGO_ENABLED=") {
			if e != "CGO_ENABLED=1" {
				t.Fatalf("expected CGO_ENABLED=1, got %s", e)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("CGO_ENABLED not found in command environment")
	}
}

// TestNewDefaultInstallerWithOptionsNilCGOEnabled verifies that a nil cgoenabled inherits CGO_ENABLED from the environment.
func TestNewDefaultInstallerWithOptionsNilCGOEnabled(t *testing.T) {
	t.Setenv("CGO_ENABLED", "0")
	inst := NewDefaultInstallerWithOptions("", nil)
	cmd := inst.buildInstallCmd("github.com/example/tool", "v1.0.0")

	if slices.Contains(cmd.Env, "CGO_ENABLED=0") {
		return
	}
	t.Fatal("expected inherited CGO_ENABLED=0 not found in command environment")
}

func TestConfigurePrivateEnv(t *testing.T) {
	const token = "secret-token"
	env, err := configurePrivateEnv(
		[]string{
			"PATH=/usr/bin",
			"GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=credential.helper",
			"GIT_CONFIG_VALUE_0=osxkeychain",
		},
		map[string]string{
			"GOPRIVATE": "github.com/acme/*",
			"GONOPROXY": "github.com/acme/*",
			"GONOSUMDB": "github.com/acme/*",
		},
		InstallOptions{
			PrivateModule: "github.com/JaredHatfield/hello-go-release",
			GitHubToken:   token,
		},
	)
	if err != nil {
		t.Fatalf("configurePrivateEnv returned error: %v", err)
	}

	wantPrivate := "github.com/acme/*,github.com/JaredHatfield/hello-go-release"
	for _, name := range []string{"GOPRIVATE", "GONOPROXY", "GONOSUMDB"} {
		value, ok := getEnv(env, name)
		if !ok || value != wantPrivate {
			t.Fatalf("%s = %q, want %q", name, value, wantPrivate)
		}
	}
	if value, _ := getEnv(env, "GIT_CONFIG_COUNT"); value != "3" {
		t.Fatalf("GIT_CONFIG_COUNT = %q, want 3", value)
	}
	if value, _ := getEnv(env, "GIT_CONFIG_KEY_1"); value != "http.https://github.com/JaredHatfield/hello-go-release.extraheader" {
		t.Fatalf("unexpected Git config key: %q", value)
	}
	if value, _ := getEnv(env, "GIT_CONFIG_KEY_2"); value != "http.https://github.com/JaredHatfield/hello-go-release.git.extraheader" {
		t.Fatalf("unexpected .git config key: %q", value)
	}
	header, ok := getEnv(env, "GIT_CONFIG_VALUE_1")
	if !ok || !strings.HasPrefix(header, "AUTHORIZATION: basic ") {
		t.Fatalf("unexpected Git authorization header: %q", header)
	}
	if strings.Contains(header, token) {
		t.Fatal("authorization header contains the unencoded token")
	}
	if value, _ := getEnv(env, "GIT_TERMINAL_PROMPT"); value != "0" {
		t.Fatalf("GIT_TERMINAL_PROMPT = %q, want 0", value)
	}
}

func TestConfigurePrivateEnvRequiresToken(t *testing.T) {
	_, err := configurePrivateEnv(nil, nil, InstallOptions{PrivateModule: "github.com/acme/tool"})
	if err == nil || !strings.Contains(err.Error(), "requires authentication") {
		t.Fatalf("expected authentication error, got %v", err)
	}
}

func TestConfigurePrivateEnvDoesNotDuplicateModule(t *testing.T) {
	const module = "github.com/acme/tool"
	env, err := configurePrivateEnv(nil, map[string]string{
		"GOPRIVATE": module,
		"GONOPROXY": module,
		"GONOSUMDB": module,
	}, InstallOptions{PrivateModule: module, GitHubToken: "token"})
	if err != nil {
		t.Fatalf("configurePrivateEnv returned error: %v", err)
	}
	for _, name := range []string{"GOPRIVATE", "GONOPROXY", "GONOSUMDB"} {
		if value, _ := getEnv(env, name); value != module {
			t.Fatalf("%s = %q, want %q", name, value, module)
		}
	}
}

func TestRedactToken(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("x-access-token:token-secret"))
	got := redactToken("request failed for token-secret and "+encoded, "token-secret")
	if got != "request failed for [REDACTED] and [REDACTED]" {
		t.Fatalf("unexpected redacted value %q", got)
	}
}
