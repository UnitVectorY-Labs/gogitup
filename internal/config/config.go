package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// App represents a configured application.
type App struct {
	Name string `yaml:"name"`
}

// Config represents the gogitup configuration file.
type Config struct {
	Apps       []App `yaml:"apps"`
	GitHubAuth bool  `yaml:"github_auth"`
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
	return &cfg, nil
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
	if HasApp(cfg, name) {
		return errors.New("app already exists: " + name)
	}
	cfg.Apps = append(cfg.Apps, App{Name: name})
	return nil
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
