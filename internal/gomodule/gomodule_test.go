package gomodule

import (
	"strings"
	"testing"
)

func TestParseUpdateAvailable(t *testing.T) {
	result, err := ParseUpdate([]byte(`{
		"Path":"golang.org/x/vuln",
		"Version":"v1.2.3",
		"Update":{"Path":"golang.org/x/vuln","Version":"v1.3.0"}
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.UpdateAvailable || result.LatestVersion != "v1.3.0" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestParseUpdateNotAvailable(t *testing.T) {
	result, err := ParseUpdate([]byte(`{"Path":"golang.org/x/vuln","Version":"v1.3.0"}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UpdateAvailable || result.LatestVersion != "v1.3.0" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestParseUpdateDoesNotInferDowngrade(t *testing.T) {
	const version = "v1.3.1-0.20260508232743-57fb27ec3243"
	result, err := ParseUpdate([]byte(`{
		"Path":"golang.org/x/vuln",
		"Version":"v1.3.1-0.20260508232743-57fb27ec3243"
	}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.UpdateAvailable || result.LatestVersion != version {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestParseUpdateRequiresVersion(t *testing.T) {
	if _, err := ParseUpdate([]byte(`{"Path":"golang.org/x/vuln"}`)); err == nil {
		t.Fatal("expected an error for missing version")
	}
}

func TestParseUpdateRequiresUpdateVersion(t *testing.T) {
	if _, err := ParseUpdate([]byte(`{
		"Path":"golang.org/x/vuln",
		"Version":"v1.2.3",
		"Update":{"Path":"golang.org/x/vuln"}
	}`)); err == nil {
		t.Fatal("expected an error for missing update version")
	}
}

func TestBuildListCmdUsesConfiguredGOPROXY(t *testing.T) {
	t.Setenv("GOPROXY", "https://environment.example.com")

	resolver := NewDefaultResolverWithGOPROXY("https://proxy.example.com,direct")
	cmd := resolver.buildListCmd("golang.org/x/vuln", "v1.2.3")

	if got := cmd.Args[len(cmd.Args)-1]; got != "golang.org/x/vuln@v1.2.3" {
		t.Fatalf("unexpected module argument %q", got)
	}
	if !containsArg(cmd.Args, "-u") {
		t.Fatalf("expected -u in command arguments: %v", cmd.Args)
	}

	var values []string
	for _, entry := range cmd.Env {
		if strings.HasPrefix(entry, "GOPROXY=") {
			values = append(values, entry)
		}
	}
	if len(values) != 1 || values[0] != "GOPROXY=https://proxy.example.com,direct" {
		t.Fatalf("unexpected GOPROXY environment: %v", values)
	}
}

func TestBuildListCmdInheritsGOPROXY(t *testing.T) {
	t.Setenv("GOPROXY", "https://environment.example.com")

	cmd := NewDefaultResolver().buildListCmd("golang.org/x/vuln", "v1.2.3")

	found := false
	for _, entry := range cmd.Env {
		if entry == "GOPROXY=https://environment.example.com" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected GOPROXY to be inherited")
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}
