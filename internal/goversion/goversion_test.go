package goversion

import (
	"errors"
	"testing"
)

func TestIsGitHubRepo(t *testing.T) {
	if !IsGitHubRepo("github.com/owner/repo") {
		t.Fatalf("expected true for github.com/owner/repo")
	}
	if !IsGitHubRepo("github.com/UnitVectorY-Labs/gogitup") {
		t.Fatalf("expected true for github.com/UnitVectorY-Labs/gogitup")
	}
	if IsGitHubRepo("gitlab.com/owner/repo") {
		t.Fatalf("expected false for gitlab.com/owner/repo")
	}
	if IsGitHubRepo("") {
		t.Fatalf("expected false for empty string")
	}
	if IsGitHubRepo("github.com") {
		t.Fatalf("expected false for github.com without trailing slash")
	}
}

func TestParseGitHubRepo(t *testing.T) {
	owner, repo, err := ParseGitHubRepo("github.com/UnitVectorY-Labs/gogitup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "UnitVectorY-Labs" {
		t.Fatalf("expected owner UnitVectorY-Labs, got %s", owner)
	}
	if repo != "gogitup" {
		t.Fatalf("expected repo gogitup, got %s", repo)
	}
}

func TestParseGitHubRepoWithSubpath(t *testing.T) {
	owner, repo, err := ParseGitHubRepo("github.com/owner/repo/subpath")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "owner" {
		t.Fatalf("expected owner owner, got %s", owner)
	}
	if repo != "repo" {
		t.Fatalf("expected repo repo, got %s", repo)
	}
}

func TestParseGitHubRepoInvalidPaths(t *testing.T) {
	_, _, err := ParseGitHubRepo("gitlab.com/owner/repo")
	if err == nil {
		t.Fatalf("expected error for non-GitHub path")
	}

	_, _, err = ParseGitHubRepo("github.com/")
	if err == nil {
		t.Fatalf("expected error for missing owner and repo")
	}

	_, _, err = ParseGitHubRepo("github.com/owner")
	if err == nil {
		t.Fatalf("expected error for missing repo")
	}

	_, _, err = ParseGitHubRepo("")
	if err == nil {
		t.Fatalf("expected error for empty string")
	}
}

func TestParseVersionJSON(t *testing.T) {
	jsonData := []byte(`[{
		"Path": "github.com/UnitVectorY-Labs/gogitup/cmd/gogitup",
		"Main": {
			"Path": "github.com/UnitVectorY-Labs/gogitup",
			"Version": "v0.1.0"
		},
		"GoVersion": "go1.23.0"
	}]`)

	info, err := ParseVersionJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Path != "github.com/UnitVectorY-Labs/gogitup" {
		t.Fatalf("expected path github.com/UnitVectorY-Labs/gogitup, got %s", info.Path)
	}
	if info.PackagePath != "github.com/UnitVectorY-Labs/gogitup/cmd/gogitup" {
		t.Fatalf("expected package path github.com/UnitVectorY-Labs/gogitup/cmd/gogitup, got %s", info.PackagePath)
	}
	if info.Version != "v0.1.0" {
		t.Fatalf("expected version v0.1.0, got %s", info.Version)
	}
	if info.GoVersion != "go1.23.0" {
		t.Fatalf("expected GoVersion go1.23.0, got %s", info.GoVersion)
	}
}

func TestParseVersionJSONSingleObject(t *testing.T) {
	jsonData := []byte(`{
		"GoVersion": "go1.25.7",
		"Path": "github.com/UnitVectorY-Labs/bulkfilepr",
		"Main": {
			"Path": "github.com/UnitVectorY-Labs/bulkfilepr",
			"Version": "v0.2.2"
		}
	}`)

	info, err := ParseVersionJSON(jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Path != "github.com/UnitVectorY-Labs/bulkfilepr" {
		t.Fatalf("expected path github.com/UnitVectorY-Labs/bulkfilepr, got %s", info.Path)
	}
	if info.Version != "v0.2.2" {
		t.Fatalf("expected version v0.2.2, got %s", info.Version)
	}
	if info.GoVersion != "go1.25.7" {
		t.Fatalf("expected GoVersion go1.25.7, got %s", info.GoVersion)
	}
}

func TestParseVersionJSONEmptyArray(t *testing.T) {
	_, err := ParseVersionJSON([]byte(`[]`))
	if err == nil {
		t.Fatalf("expected error for empty array")
	}
}

func TestParseVersionJSONMissingModule(t *testing.T) {
	jsonData := []byte(`[{
		"Path": "/usr/local/bin/something",
		"Main": {
			"Path": "",
			"Version": ""
		},
		"GoVersion": "go1.23.0"
	}]`)

	_, err := ParseVersionJSON(jsonData)
	if err == nil {
		t.Fatalf("expected error for missing module info")
	}
}

func TestParseVersionJSONInvalid(t *testing.T) {
	_, err := ParseVersionJSON([]byte(`not json`))
	if err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

// mockRunner is a mock implementation of Runner for testing.
type mockRunner struct {
	info *Info
	err  error
}

func (m *mockRunner) GetInfo(binaryName string) (*Info, error) {
	return m.info, m.err
}

func TestMockRunnerSatisfiesInterface(t *testing.T) {
	var r Runner = &mockRunner{
		info: &Info{
			Path:      "github.com/UnitVectorY-Labs/gogitup",
			Version:   "v0.1.0",
			GoVersion: "go1.23.0",
		},
	}

	info, err := r.GetInfo("gogitup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Path != "github.com/UnitVectorY-Labs/gogitup" {
		t.Fatalf("expected path github.com/UnitVectorY-Labs/gogitup, got %s", info.Path)
	}
	if info.Version != "v0.1.0" {
		t.Fatalf("expected version v0.1.0, got %s", info.Version)
	}
}

func TestMockRunnerError(t *testing.T) {
	var r Runner = &mockRunner{
		err: errors.New("binary not found: missing"),
	}

	_, err := r.GetInfo("missing")
	if err == nil {
		t.Fatalf("expected error from mock runner")
	}
}

func TestIsToolchainNewer(t *testing.T) {
	tests := []struct {
		name      string
		active    string
		builtWith string
		want      bool
	}{
		{name: "newer patch", active: "go1.25.7", builtWith: "go1.25.6", want: true},
		{name: "newer minor", active: "go1.26.0", builtWith: "go1.25.7", want: true},
		{name: "same", active: "go1.25.7", builtWith: "go1.25.7"},
		{name: "older", active: "go1.24.0", builtWith: "go1.25.7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsToolchainNewer(tt.active, tt.builtWith)
			if err != nil {
				t.Fatalf("IsToolchainNewer returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsToolchainNewer(%q, %q) = %v, want %v", tt.active, tt.builtWith, got, tt.want)
			}
		})
	}
}

func TestIsToolchainNewerRejectsInvalidVersions(t *testing.T) {
	if _, err := IsToolchainNewer("1.25.7", "go1.25.6"); err == nil {
		t.Fatal("expected invalid active version error")
	}
	if _, err := IsToolchainNewer("go1.25.7", "unknown"); err == nil {
		t.Fatal("expected invalid embedded version error")
	}
}
