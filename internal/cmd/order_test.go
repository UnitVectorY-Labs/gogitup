package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

// Deliberately include mixed case and names that differ only by case.
func unorderedApps() []config.App {
	return []config.App{{Name: "zulu"}, {Name: "beta"}, {Name: "alpha"}, {Name: "Beta"}}
}

var alphabeticalNames = []string{"alpha", "Beta", "beta", "zulu"}

type orderedRunner struct {
	calls []string
	fail  bool
}

func (r *orderedRunner) GetInfo(name string) (*goversion.Info, error) {
	r.calls = append(r.calls, name)
	if r.fail {
		return nil, errors.New("missing binary")
	}
	return &goversion.Info{Path: "github.com/acme/" + name, Version: "v1.0.0", GoVersion: "go1.25.6"}, nil
}

func assertOrder(t *testing.T, got []string) {
	t.Helper()
	if !reflect.DeepEqual(got, alphabeticalNames) {
		t.Fatalf("order=%v want=%v", got, alphabeticalNames)
	}
}

func assertOutputOrder(t *testing.T, text string) {
	t.Helper()
	previous := -1
	for _, name := range alphabeticalNames {
		index := strings.Index(text, name)
		if index <= previous {
			t.Fatalf("%s out of order in %q", name, text)
		}
		previous = index
	}
}

func TestListAlphabeticalProcessingAndOutput(t *testing.T) {
	for _, fail := range []bool{false, true} {
		apps := unorderedApps()
		original := append([]config.App(nil), apps...)
		runner := &orderedRunner{fail: fail}
		entries := collectListEntries(apps, runner)
		assertOrder(t, runner.calls)
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name)
		}
		assertOrder(t, names)
		var table bytes.Buffer
		printListTable(&table, entries, "")
		assertOutputOrder(t, table.String())
		data, err := json.Marshal(entries)
		if err != nil {
			t.Fatal(err)
		}
		assertOutputOrder(t, string(data))
		if !reflect.DeepEqual(apps, original) {
			t.Fatal("list changed config order")
		}
	}
}

func TestCheckAlphabeticalProcessingAndOutput(t *testing.T) {
	for _, mode := range []string{"fresh", "cached", "force", "errors"} {
		t.Run(mode, func(t *testing.T) {
			apps := unorderedApps()
			original := append([]config.App(nil), apps...)
			runner := &orderedRunner{fail: mode == "errors"}
			c := &cache.Cache{Entries: map[string]cache.Entry{}}
			if mode == "cached" || mode == "force" {
				for _, app := range apps {
					cache.SetForInstalledVersion(c, app.Name, "v1.0.0", "v2.0.0")
				}
			}
			var lookups []string
			var stderr bytes.Buffer
			entries, failed := collectCheckEntries(apps, c, mode == "force", checkDependencies{
				runner: runner,
				githubForApp: func(app config.App) (github.Client, string, error) {
					lookups = append(lookups, app.Name)
					return &stubGitHubClient{releases: map[string]string{"acme/" + app.Name: "v2.0.0"}}, "", nil
				},
				errOut: &output.Writer{Out: &stderr},
			})
			assertOrder(t, runner.calls)
			if mode == "fresh" || mode == "force" {
				assertOrder(t, lookups)
			} else if len(lookups) != 0 {
				t.Fatalf("unexpected lookups: %v", lookups)
			}
			if failed != (mode == "errors") {
				t.Fatalf("failed=%v", failed)
			}
			if failed {
				assertOutputOrder(t, stderr.String())
			}
			var names []string
			for _, entry := range entries {
				names = append(names, entry.Name)
			}
			assertOrder(t, names)
			data, err := json.Marshal(entries)
			if err != nil {
				t.Fatal(err)
			}
			assertOutputOrder(t, string(data))
			if !reflect.DeepEqual(apps, original) {
				t.Fatal("check changed config order")
			}
		})
	}
}

func TestUpgradeAlphabeticalActionsAndOutput(t *testing.T) {
	for _, mode := range []string{"upgrade", "verbose", "rebuild", "dry-run", "dry-run-rebuild", "errors"} {
		t.Run(mode, func(t *testing.T) {
			apps := unorderedApps()
			original := append([]config.App(nil), apps...)
			runner := &orderedRunner{fail: mode == "errors"}
			inst := &stubInstaller{}
			var stdout, stderr bytes.Buffer
			var lookups []string
			rebuild := strings.Contains(mode, "rebuild")
			dryRun := strings.HasPrefix(mode, "dry-run")
			updated, failed := runUpgradeApps(&config.Config{Apps: apps}, &cache.Cache{Entries: map[string]cache.Entry{}}, upgradeOptions{
				Verbose: mode == "verbose", RebuildWithNewerGo: rebuild, DryRun: dryRun,
			}, upgradeDependencies{
				runner: runner, installer: inst,
				currentGo: func() (string, error) { return "go1.25.7", nil },
				githubForApp: func(app config.App) (github.Client, string, error) {
					lookups = append(lookups, app.Name)
					version := "v2.0.0"
					if rebuild || mode == "verbose" {
						version = "v1.0.0"
					}
					return &stubGitHubClient{releases: map[string]string{"acme/" + app.Name: version}}, "", nil
				},
				out: &output.Writer{Out: &stdout}, errOut: &output.Writer{Out: &stderr},
			})
			assertOrder(t, runner.calls)
			if mode == "errors" {
				if !failed || updated != 0 {
					t.Fatalf("updated=%d failed=%v", updated, failed)
				}
				assertOutputOrder(t, stderr.String())
			} else {
				assertOrder(t, lookups)
				assertOutputOrder(t, stdout.String())
				want := 4
				if mode == "verbose" {
					want = 0
				}
				if failed || updated != want {
					t.Fatalf("updated=%d failed=%v", updated, failed)
				}
			}
			if !dryRun && mode != "verbose" && mode != "errors" {
				var names []string
				for _, call := range inst.calls {
					names = append(names, strings.TrimPrefix(call.modulePath, "github.com/acme/"))
				}
				assertOrder(t, names)
			} else if len(inst.calls) != 0 {
				t.Fatal("unexpected installation")
			}
			if !reflect.DeepEqual(apps, original) {
				t.Fatal("upgrade changed config order")
			}
		})
	}
}
