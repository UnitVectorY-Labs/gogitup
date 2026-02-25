package cache

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultTTL is the default time-to-live for cache entries.
const DefaultTTL = 24 * time.Hour

// Entry represents a cached version check result for an application.
type Entry struct {
	LatestVersion string    `yaml:"latest_version"`
	CheckedAt     time.Time `yaml:"checked_at"`
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
	c.Entries[name] = Entry{
		LatestVersion: version,
		CheckedAt:     time.Now(),
	}
}

// IsExpired checks if a cache entry is older than the given TTL.
func IsExpired(entry Entry, ttl time.Duration) bool {
	return time.Since(entry.CheckedAt) > ttl
}

// Remove removes the cache entry for the given app name.
func Remove(c *Cache, name string) {
	delete(c.Entries, name)
}
