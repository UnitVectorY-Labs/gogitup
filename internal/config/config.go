package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

// App represents a configured application.
type App struct {
	Name        string `yaml:"name"`
	InstallPath string `yaml:"install_path,omitempty"`
	Private     bool   `yaml:"private,omitempty"`
	GitHubUser  string `yaml:"github_user,omitempty"`
}

// Config represents the gogitup configuration file.
type Config struct {
	Apps       []App  `yaml:"apps"`
	GitHubAuth bool   `yaml:"github_auth"`
	GOPROXY    string `yaml:"goproxy,omitempty"`
	CGOEnabled *bool  `yaml:"cgo_enabled,omitempty"`
}

// DefaultPath returns the default config file path (~/.gogitup).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".gogitup")
	}
	return filepath.Join(home, ".gogitup")
}

// Load reads and parses the config file at the given path.
// If the file does not exist, an empty Config is returned without error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	for _, app := range cfg.Apps {
		if err := ValidateGitHubUser(app.GitHubUser); err != nil {
			return nil, fmt.Errorf("app %q: %w", app.Name, err)
		}
	}
	return &cfg, nil
}

// ValidateGitHubUser permits an omitted account but rejects whitespace and controls.
func ValidateGitHubUser(user string) error {
	if strings.ContainsFunc(user, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return errors.New("github_user must not contain whitespace or control characters")
	}
	return nil
}

// Save writes the config to the given file path.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// AddApp adds an app to the config. Returns an error if the app already exists.
func AddApp(cfg *Config, name string) error {
	return AddAppWithInstallPath(cfg, name, "")
}

// AddAppWithInstallPath adds an app with an optional Go package path used for
// future upgrades. Returns an error if the app already exists.
func AddAppWithInstallPath(cfg *Config, name, installPath string) error {
	return AddAppWithInstallOptions(cfg, name, installPath, false)
}

// AddAppWithInstallOptions adds an app with the values needed for future
// upgrades. Returns an error if the app already exists.
func AddAppWithInstallOptions(cfg *Config, name, installPath string, private bool) error {
	return RegisterApp(cfg, App{Name: name, InstallPath: installPath, Private: private})
}

// RegisterApp adds an application's persistent options without changing existing entries.
func RegisterApp(cfg *Config, app App) error {
	if err := ValidateGitHubUser(app.GitHubUser); err != nil {
		return err
	}
	if HasApp(cfg, app.Name) {
		return errors.New("app already exists: " + app.Name)
	}
	cfg.Apps = append(cfg.Apps, app)
	return nil
}

// HasPrivateApps reports whether any tracked application needs authenticated
// private GitHub access.
func HasPrivateApps(cfg *Config) bool {
	for _, app := range cfg.Apps {
		if app.Private {
			return true
		}
	}
	return false
}

// RemoveApp removes an app from the config. Returns an error if the app is not found.
func RemoveApp(cfg *Config, name string) error {
	for i, app := range cfg.Apps {
		if app.Name == name {
			cfg.Apps = append(cfg.Apps[:i], cfg.Apps[i+1:]...)
			return nil
		}
	}
	return errors.New("app not found: " + name)
}

// HasApp checks if an app exists in the config.
func HasApp(cfg *Config, name string) bool {
	for _, app := range cfg.Apps {
		if app.Name == name {
			return true
		}
	}
	return false
}
