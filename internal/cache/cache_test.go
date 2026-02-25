package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLoadNonExistentFile(t *testing.T) {
	c, err := Load("/tmp/gogitup_test_nonexistent_cache")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(c.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(c.Entries))
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gogitup.cache")

	original := &Cache{
		Entries: map[string]Entry{
			"app1": {LatestVersion: "v1.2.3", CheckedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
			"app2": {LatestVersion: "v2.0.0", CheckedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)},
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	if len(loaded.Entries) != len(original.Entries) {
		t.Fatalf("expected %d entries, got %d", len(original.Entries), len(loaded.Entries))
	}
	for name, orig := range original.Entries {
		got, ok := loaded.Entries[name]
		if !ok {
			t.Fatalf("expected entry for %s", name)
		}
		if got.LatestVersion != orig.LatestVersion {
			t.Fatalf("expected version %s for %s, got %s", orig.LatestVersion, name, got.LatestVersion)
		}
		if !got.CheckedAt.Equal(orig.CheckedAt) {
			t.Fatalf("expected checked_at %v for %s, got %v", orig.CheckedAt, name, got.CheckedAt)
		}
	}
}

func TestGetSet(t *testing.T) {
	c := &Cache{Entries: make(map[string]Entry)}

	_, ok := Get(c, "myapp")
	if ok {
		t.Fatal("expected Get to return false for missing entry")
	}

	Set(c, "myapp", "v1.0.0")

	entry, ok := Get(c, "myapp")
	if !ok {
		t.Fatal("expected Get to return true after Set")
	}
	if entry.LatestVersion != "v1.0.0" {
		t.Fatalf("expected version v1.0.0, got %s", entry.LatestVersion)
	}
	if time.Since(entry.CheckedAt) > time.Second {
		t.Fatal("expected CheckedAt to be recent")
	}
}

func TestIsExpired(t *testing.T) {
	recent := Entry{
		LatestVersion: "v1.0.0",
		CheckedAt:     time.Now(),
	}
	if IsExpired(recent, DefaultTTL) {
		t.Fatal("expected recent entry to not be expired")
	}

	old := Entry{
		LatestVersion: "v1.0.0",
		CheckedAt:     time.Now().Add(-48 * time.Hour),
	}
	if !IsExpired(old, DefaultTTL) {
		t.Fatal("expected old entry to be expired")
	}

	borderline := Entry{
		LatestVersion: "v1.0.0",
		CheckedAt:     time.Now().Add(-23 * time.Hour),
	}
	if IsExpired(borderline, DefaultTTL) {
		t.Fatal("expected borderline entry to not be expired")
	}
}

func TestRemove(t *testing.T) {
	c := &Cache{
		Entries: map[string]Entry{
			"app1": {LatestVersion: "v1.0.0", CheckedAt: time.Now()},
			"app2": {LatestVersion: "v2.0.0", CheckedAt: time.Now()},
		},
	}

	Remove(c, "app1")

	if _, ok := Get(c, "app1"); ok {
		t.Fatal("expected app1 to be removed")
	}
	if _, ok := Get(c, "app2"); !ok {
		t.Fatal("expected app2 to still exist")
	}
}
