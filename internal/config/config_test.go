package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNonExistentFile(t *testing.T) {
	cfg, err := Load("/tmp/gogitup_test_nonexistent_file")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Apps) != 0 {
		t.Fatalf("expected empty apps, got %d", len(cfg.Apps))
	}
	if cfg.GitHubAuth {
		t.Fatal("expected GitHubAuth to be false")
	}
}

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gogitup")

	content := []byte("apps:\n  - name: \"app1\"\n  - name: \"app2\"\ngithub_auth: true\n")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Apps) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "app1" {
		t.Fatalf("expected app1, got %s", cfg.Apps[0].Name)
	}
	if cfg.Apps[1].Name != "app2" {
		t.Fatalf("expected app2, got %s", cfg.Apps[1].Name)
	}
	if !cfg.GitHubAuth {
		t.Fatal("expected GitHubAuth to be true")
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gogitup")

	original := &Config{
		Apps:       []App{{Name: "myapp"}, {Name: "otherapp"}},
		GitHubAuth: true,
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loaded.Apps) != len(original.Apps) {
		t.Fatalf("expected %d apps, got %d", len(original.Apps), len(loaded.Apps))
	}
	for i, app := range loaded.Apps {
		if app.Name != original.Apps[i].Name {
			t.Fatalf("expected app %s, got %s", original.Apps[i].Name, app.Name)
		}
	}
	if loaded.GitHubAuth != original.GitHubAuth {
		t.Fatalf("expected GitHubAuth %v, got %v", original.GitHubAuth, loaded.GitHubAuth)
	}
}

func TestAddApp(t *testing.T) {
	cfg := &Config{}

	if err := AddApp(cfg, "newapp"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "newapp" {
		t.Fatalf("expected newapp, got %s", cfg.Apps[0].Name)
	}

	// Adding duplicate should return error
	if err := AddApp(cfg, "newapp"); err == nil {
		t.Fatal("expected error for duplicate app, got nil")
	}
}

func TestRemoveApp(t *testing.T) {
	cfg := &Config{
		Apps: []App{{Name: "app1"}, {Name: "app2"}},
	}

	if err := RemoveApp(cfg, "app1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(cfg.Apps))
	}
	if cfg.Apps[0].Name != "app2" {
		t.Fatalf("expected app2, got %s", cfg.Apps[0].Name)
	}

	// Removing non-existent should return error
	if err := RemoveApp(cfg, "nonexistent"); err == nil {
		t.Fatal("expected error for non-existent app, got nil")
	}
}

func TestHasApp(t *testing.T) {
	cfg := &Config{
		Apps: []App{{Name: "app1"}},
	}

	if !HasApp(cfg, "app1") {
		t.Fatal("expected HasApp to return true for app1")
	}
	if HasApp(cfg, "app2") {
		t.Fatal("expected HasApp to return false for app2")
	}
}
