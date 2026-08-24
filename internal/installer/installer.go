package installer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// InstallOptions configures an individual go install invocation.
type InstallOptions struct {
	PrivateModule string
	GitHubToken   string
}

// Installer defines the interface for installing Go modules.
type Installer interface {
	Install(modulePath string, version string, options InstallOptions) (string, error)
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

func loadPrivateGoEnv() (map[string]string, error) {
	cmd := exec.Command("go", "env", "-json", "GOPRIVATE", "GONOPROXY", "GONOSUMDB")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("go env failed: %w\n%s", err, string(out))
	}

	values := make(map[string]string)
	if err := json.Unmarshal(out, &values); err != nil {
		return nil, fmt.Errorf("failed to parse go env output: %w", err)
	}
	return values, nil
}

func configurePrivateEnv(env []string, goEnv map[string]string, options InstallOptions) ([]string, error) {
	if options.PrivateModule == "" {
		return env, nil
	}
	if options.GitHubToken == "" {
		return nil, fmt.Errorf("private GitHub access requires authentication; set GITHUB_TOKEN or run 'gh auth login'")
	}
	repositoryRoot, err := githubRepositoryRoot(options.PrivateModule)
	if err != nil {
		return nil, err
	}

	for _, name := range []string{"GOPRIVATE", "GONOPROXY", "GONOSUMDB"} {
		env = setEnv(env, name, appendListValue(goEnv[name], options.PrivateModule))
	}

	count := 0
	if value, ok := getEnv(env, "GIT_CONFIG_COUNT"); ok && value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return nil, fmt.Errorf("invalid GIT_CONFIG_COUNT value %q", value)
		}
		count = parsed
	}

	credentials := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + options.GitHubToken))
	repositoryURL := "https://" + repositoryRoot
	env = setEnv(env, "GIT_CONFIG_COUNT", strconv.Itoa(count+2))
	env = setEnv(env, fmt.Sprintf("GIT_CONFIG_KEY_%d", count), "http."+repositoryURL+".extraheader")
	env = setEnv(env, fmt.Sprintf("GIT_CONFIG_VALUE_%d", count), "AUTHORIZATION: basic "+credentials)
	env = setEnv(env, fmt.Sprintf("GIT_CONFIG_KEY_%d", count+1), "http."+repositoryURL+".git.extraheader")
	env = setEnv(env, fmt.Sprintf("GIT_CONFIG_VALUE_%d", count+1), "AUTHORIZATION: basic "+credentials)
	env = setEnv(env, "GIT_TERMINAL_PROMPT", "0")
	return env, nil
}

func githubRepositoryRoot(modulePath string) (string, error) {
	parts := strings.Split(strings.TrimSuffix(modulePath, "/"), "/")
	if len(parts) < 3 || parts[0] != "github.com" || parts[1] == "" || parts[2] == "" {
		return "", fmt.Errorf("private module %q is not a valid github.com repository path", modulePath)
	}
	return strings.Join(parts[:3], "/"), nil
}

func appendListValue(current, value string) string {
	for _, item := range strings.Split(current, ",") {
		if strings.TrimSpace(item) == value {
			return current
		}
	}
	if current == "" {
		return value
	}
	return current + "," + value
}

func getEnv(env []string, name string) (string, bool) {
	prefix := name + "="
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], prefix) {
			return strings.TrimPrefix(env[i], prefix), true
		}
	}
	return "", false
}

func setEnv(env []string, name, value string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered, prefix+value)
}

func redactToken(value, token string) string {
	if token == "" {
		return value
	}
	value = strings.ReplaceAll(value, token, "[REDACTED]")
	credentials := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))
	return strings.ReplaceAll(value, credentials, "[REDACTED]")
}

// Install runs "go install {modulePath}@{version}" and returns the combined output.
func (d *DefaultInstaller) Install(modulePath string, version string, options InstallOptions) (string, error) {
	cmd := d.buildInstallCmd(modulePath, version)
	if options.PrivateModule != "" {
		goEnv, err := loadPrivateGoEnv()
		if err != nil {
			return "", err
		}
		cmd.Env, err = configurePrivateEnv(cmd.Env, goEnv, options)
		if err != nil {
			return "", err
		}
	}

	out, err := cmd.CombinedOutput()
	output := redactToken(string(out), options.GitHubToken)
	if err != nil {
		return output, fmt.Errorf("go install %s@%s failed: %w\n%s", modulePath, version, err, output)
	}
	return output, nil
}
