package cache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultTTL is the default time-to-live for cache entries.
const DefaultTTL = 24 * time.Hour

// Entry represents a cached version check result for an application.
type Entry struct {
	LatestVersion    string    `yaml:"latest_version"`
	InstalledVersion string    `yaml:"installed_version,omitempty"`
	CheckedAt        time.Time `yaml:"checked_at"`
	GitHubUser       string    `yaml:"github_user,omitempty"`
	Private          bool      `yaml:"private,omitempty"`
}

// Cache represents the gogitup cache file.
type Cache struct {
	Entries map[string]Entry `yaml:"entries"`
}

// DefaultPath returns the default cache file path (~/.gogitup.cache).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".gogitup.cache")
	}
	return filepath.Join(home, ".gogitup.cache")
}

// Load reads and parses the cache file at the given path.
// If the file does not exist, an empty Cache is returned without error.
func Load(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Cache{Entries: make(map[string]Entry)}, nil
		}
		return nil, err
	}

	var c Cache
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Entries == nil {
		c.Entries = make(map[string]Entry)
	}
	return &c, nil
}

// Save writes the cache to the given file path.
func Save(path string, c *Cache) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// Get returns the cache entry for the given app name and whether it was found.
func Get(c *Cache, name string) (Entry, bool) {
	entry, ok := c.Entries[name]
	return entry, ok
}

// Set sets the cache entry for the given app name with the current time.
func Set(c *Cache, name string, version string) {
	SetForInstalledVersion(c, name, "", version)
}

// SetForInstalledVersion caches a version check for a specific installed version.
func SetForInstalledVersion(c *Cache, name, installedVersion, latestVersion string) {
	SetForApp(c, name, installedVersion, latestVersion, "", false)
}

// SetForApp binds a version result to the app's configured account and privacy.
func SetForApp(c *Cache, name, installedVersion, latestVersion, user string, private bool) {
	c.Entries[name] = Entry{
		LatestVersion:    latestVersion,
		InstalledVersion: installedVersion,
		CheckedAt:        time.Now(),
		GitHubUser:       user,
		Private:          private,
	}
}

// Matches reports whether a result applies to the current app configuration.
func (e Entry) Matches(installedVersion, user string, private bool) bool {
	return e.InstalledVersion == installedVersion && strings.EqualFold(e.GitHubUser, user) && e.Private == private
}

// IsExpired checks if a cache entry is older than the given TTL.
func IsExpired(entry Entry, ttl time.Duration) bool {
	return time.Since(entry.CheckedAt) > ttl
}

// Remove removes the cache entry for the given app name.
func Remove(c *Cache, name string) {
	delete(c.Entries, name)
}
