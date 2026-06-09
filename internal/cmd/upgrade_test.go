package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/gomodule"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type stubRunner struct {
	infos map[string]*goversion.Info
	errs  map[string]error
}

func (s *stubRunner) GetInfo(binaryName string) (*goversion.Info, error) {
	if err, ok := s.errs[binaryName]; ok {
		return nil, err
	}

	info, ok := s.infos[binaryName]
	if !ok {
		return nil, errors.New("info not found")
	}

	return info, nil
}

type stubGitHubClient struct {
	releases map[string]string
	errs     map[string]error
}

func (s *stubGitHubClient) GetLatestRelease(owner, repo string) (string, error) {
	key := owner + "/" + repo
	if err, ok := s.errs[key]; ok {
		return "", err
	}

	release, ok := s.releases[key]
	if !ok {
		return "", errors.New("release not found")
	}

	return release, nil
}

type installCall struct {
	modulePath string
	version    string
}

type stubInstaller struct {
	calls []installCall
	err   error
}

type stubModuleResolver struct {
	results map[string]gomodule.Result
	errs    map[string]error
	calls   []moduleCheckCall
}

type moduleCheckCall struct {
	modulePath       string
	installedVersion string
}

func (s *stubModuleResolver) Check(modulePath, installedVersion string) (gomodule.Result, error) {
	s.calls = append(s.calls, moduleCheckCall{modulePath: modulePath, installedVersion: installedVersion})
	key := modulePath + "@" + installedVersion
	if err, ok := s.errs[key]; ok {
		return gomodule.Result{}, err
	}
	result, ok := s.results[key]
	if !ok {
		return gomodule.Result{}, errors.New("version result not found")
	}
	return result, nil
}

func (s *stubInstaller) Install(modulePath, version string) (string, error) {
	s.calls = append(s.calls, installCall{modulePath: modulePath, version: version})
	if s.err != nil {
		return "", s.err
	}
	return "ok", nil
}

func TestParseUpgradeOptionsVerbose(t *testing.T) {
	var stderr bytes.Buffer

	opts, err := parseUpgradeOptions([]string{"--verbose"}, &stderr)
	if err != nil {
		t.Fatalf("parseUpgradeOptions returned error: %v", err)
	}

	if !opts.Verbose {
		t.Fatalf("expected verbose option to be enabled")
	}
}

func TestRunUpgradeAppsSuppressesUpToDateEntriesByDefault(t *testing.T) {
	cfg := &config.Config{
		Apps: []config.App{
			{Name: "current"},
			{Name: "stale"},
		},
	}
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"current": {Path: "github.com/acme/current", Version: "v1.2.3"},
			"stale":   {Path: "github.com/acme/stale", Version: "v1.0.0"},
		},
	}
	ghClient := &stubGitHubClient{
		releases: map[string]string{
			"acme/current": "v1.2.3",
			"acme/stale":   "v1.1.0",
		},
	}
	installer := &stubInstaller{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	updated := runUpgradeApps(cfg, c, upgradeOptions{}, upgradeDependencies{
		runner:    runner,
		ghClient:  ghClient,
		installer: installer,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if updated != 1 {
		t.Fatalf("expected 1 updated binary, got %d", updated)
	}
	if strings.Contains(stdout.String(), "already up to date") {
		t.Fatalf("expected default output to suppress up-to-date entries, got %q", stdout.String())
	}

	progress := "Upgrading 'stale' from " + output.Green + "v1.0.0" + output.Reset + " to " + output.Cyan + "v1.1.0" + output.Reset
	if !strings.Contains(stdout.String(), progress) {
		t.Fatalf("expected progress output %q, got %q", progress, stdout.String())
	}

	success := "Upgraded 'stale' to " + output.Green + "v1.1.0" + output.Reset
	if !strings.Contains(stdout.String(), success) {
		t.Fatalf("expected success output %q, got %q", success, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}

	if len(installer.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(installer.calls))
	}
	if installer.calls[0].modulePath != "github.com/acme/stale" || installer.calls[0].version != "v1.1.0" {
		t.Fatalf("unexpected install call: %+v", installer.calls[0])
	}

	currentEntry, ok := cache.Get(c, "current")
	if !ok || currentEntry.LatestVersion != "v1.2.3" {
		t.Fatalf("expected cache to record current binary version, got %#v", currentEntry)
	}

	staleEntry, ok := cache.Get(c, "stale")
	if !ok || staleEntry.LatestVersion != "v1.1.0" {
		t.Fatalf("expected cache to record updated binary version, got %#v", staleEntry)
	}
}

func TestRunUpgradeAppsVerboseIncludesUpToDateEntries(t *testing.T) {
	cfg := &config.Config{
		Apps: []config.App{
			{Name: "current"},
			{Name: "stale"},
		},
	}
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"current": {Path: "github.com/acme/current", Version: "v1.2.3"},
			"stale":   {Path: "github.com/acme/stale", Version: "v1.0.0"},
		},
	}
	ghClient := &stubGitHubClient{
		releases: map[string]string{
			"acme/current": "v1.2.3",
			"acme/stale":   "v1.1.0",
		},
	}
	installer := &stubInstaller{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	updated := runUpgradeApps(cfg, c, upgradeOptions{Verbose: true}, upgradeDependencies{
		runner:    runner,
		ghClient:  ghClient,
		installer: installer,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if updated != 1 {
		t.Fatalf("expected 1 updated binary, got %d", updated)
	}

	infoMessage := "'current' is already up to date (" + output.Green + "v1.2.3" + output.Reset + ")"
	if !strings.Contains(stdout.String(), infoMessage) {
		t.Fatalf("expected verbose output %q, got %q", infoMessage, stdout.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}

func TestRunUpgradeAppsUsesPackagePathForNonGitHubModule(t *testing.T) {
	cfg := &config.Config{
		Apps: []config.App{{
			Name:        "govulncheck",
			InstallPath: "golang.org/x/vuln/cmd/govulncheck",
		}},
	}
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"govulncheck": {Path: "golang.org/x/vuln", Version: "v1.0.0"},
		},
	}
	resolver := &stubModuleResolver{
		results: map[string]gomodule.Result{
			"golang.org/x/vuln@v1.0.0": {LatestVersion: "v1.1.0", UpdateAvailable: true},
		},
	}
	installer := &stubInstaller{}
	var stdout, stderr bytes.Buffer

	updated := runUpgradeApps(cfg, c, upgradeOptions{}, upgradeDependencies{
		runner:    runner,
		ghClient:  &stubGitHubClient{},
		resolver:  resolver,
		installer: installer,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if updated != 1 {
		t.Fatalf("expected 1 updated binary, got %d", updated)
	}
	if len(installer.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(installer.calls))
	}
	if installer.calls[0].modulePath != "golang.org/x/vuln/cmd/govulncheck" || installer.calls[0].version != "v1.1.0" {
		t.Fatalf("unexpected install call: %+v", installer.calls[0])
	}
}

func TestRunUpgradeAppsUsesEmbeddedPackagePathWhenConfigPathIsMissing(t *testing.T) {
	cfg := &config.Config{
		Apps: []config.App{{Name: "govulncheck"}},
	}
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"govulncheck": {
				Path:        "golang.org/x/vuln",
				PackagePath: "golang.org/x/vuln/cmd/govulncheck",
				Version:     "v1.0.0",
			},
		},
	}
	resolver := &stubModuleResolver{
		results: map[string]gomodule.Result{
			"golang.org/x/vuln@v1.0.0": {LatestVersion: "v1.1.0", UpdateAvailable: true},
		},
	}
	installer := &stubInstaller{}

	updated := runUpgradeApps(cfg, c, upgradeOptions{}, upgradeDependencies{
		runner:    runner,
		ghClient:  &stubGitHubClient{},
		resolver:  resolver,
		installer: installer,
		out:       &output.Writer{Out: &bytes.Buffer{}},
		errOut:    &output.Writer{Out: &bytes.Buffer{}},
	})

	if updated != 1 {
		t.Fatalf("expected 1 updated binary, got %d", updated)
	}
	if len(installer.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(installer.calls))
	}
	if installer.calls[0].modulePath != "golang.org/x/vuln/cmd/govulncheck" || installer.calls[0].version != "v1.1.0" {
		t.Fatalf("unexpected install call: %+v", installer.calls[0])
	}
}

func TestRunUpgradeAppsDoesNotDowngradeNonGitHubModule(t *testing.T) {
	const installedVersion = "v1.3.1-0.20260508232743-57fb27ec3243"
	cfg := &config.Config{
		Apps: []config.App{{
			Name:        "govulncheck",
			InstallPath: "golang.org/x/vuln/cmd/govulncheck",
		}},
	}
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"govulncheck": {Path: "golang.org/x/vuln", Version: installedVersion},
		},
	}
	resolver := &stubModuleResolver{
		results: map[string]gomodule.Result{
			"golang.org/x/vuln@" + installedVersion: {LatestVersion: installedVersion},
		},
	}
	installer := &stubInstaller{}

	updated := runUpgradeApps(cfg, c, upgradeOptions{}, upgradeDependencies{
		runner:    runner,
		ghClient:  &stubGitHubClient{},
		resolver:  resolver,
		installer: installer,
		out:       &output.Writer{Out: &bytes.Buffer{}},
		errOut:    &output.Writer{Out: &bytes.Buffer{}},
	})

	if updated != 0 {
		t.Fatalf("expected no updates, got %d", updated)
	}
	if len(installer.calls) != 0 {
		t.Fatalf("expected no install calls, got %d", len(installer.calls))
	}
	if len(resolver.calls) != 1 || resolver.calls[0].installedVersion != installedVersion {
		t.Fatalf("unexpected resolver calls: %+v", resolver.calls)
	}
}
