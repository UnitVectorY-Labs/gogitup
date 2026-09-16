package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func TestCheckAccountCache(t *testing.T) {
	for _, tt := range []struct {
		name, cachedUser, user                    string
		cachedPrivate, private, force, wantLookup bool
	}{
		{name: "hit", cachedUser: "work", user: "work", cachedPrivate: true, private: true},
		{name: "case insensitive", cachedUser: "WORK", user: "work"},
		{name: "changed account", cachedUser: "personal", user: "work", wantLookup: true},
		{name: "removed account", cachedUser: "work", wantLookup: true},
		{name: "legacy cache", user: "work", wantLookup: true},
		{name: "privacy changed", cachedUser: "work", user: "work", private: true, wantLookup: true},
		{name: "privacy removed", cachedPrivate: true, wantLookup: true},
		{name: "force", cachedUser: "work", user: "work", force: true, wantLookup: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := &cache.Cache{Entries: map[string]cache.Entry{}}
			cache.SetForApp(c, "tool", "v1.0.0", "v1.1.0", tt.cachedUser, tt.cachedPrivate)
			calls := 0
			var stderr bytes.Buffer
			entries, failed := collectCheckEntries([]config.App{{Name: "tool", GitHubUser: tt.user, Private: tt.private}}, c, tt.force, checkDependencies{
				runner: &stubRunner{infos: map[string]*goversion.Info{"tool": {Path: "github.com/acme/tool", Version: "v1.0.0"}}},
				githubForApp: func(app config.App) (github.Client, string, error) {
					calls++
					if app.GitHubUser != tt.user || app.Private != tt.private {
						t.Fatalf("wrong account: %+v", app)
					}
					return &stubGitHubClient{releases: map[string]string{"acme/tool": "v1.2.0"}}, "token", nil
				},
				errOut: &output.Writer{Out: &stderr},
			})
			if failed || len(entries) != 1 || (calls == 1) != tt.wantLookup || stderr.Len() != 0 {
				t.Fatalf("entries=%+v calls=%d failed=%v stderr=%s", entries, calls, failed, &stderr)
			}
			want := "v1.1.0"
			if tt.wantLookup {
				want = "v1.2.0"
			}
			if entries[0].LatestVersion != want || !entries[0].UpdateAvailable {
				t.Fatalf("entry=%+v", entries[0])
			}
			if tt.wantLookup && !c.Entries["tool"].Matches("v1.0.0", tt.user, tt.private) {
				t.Fatal("cache lost auth context")
			}
		})
	}
}

func TestCheckContinuesAfterAccountFailure(t *testing.T) {
	var stderr bytes.Buffer
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	entries, failed := collectCheckEntries([]config.App{{Name: "broken", GitHubUser: "missing"}, {Name: "working", GitHubUser: "work"}}, c, false, checkDependencies{
		runner: &stubRunner{infos: map[string]*goversion.Info{
			"broken":  {Path: "github.com/acme/broken", Version: "v1.0.0"},
			"working": {Path: "github.com/acme/working", Version: "v1.0.0"},
		}},
		githubForApp: func(app config.App) (github.Client, string, error) {
			if app.GitHubUser == "missing" {
				return nil, "", errors.New("account not logged in")
			}
			return &stubGitHubClient{releases: map[string]string{"acme/working": "v2.0.0"}}, "work-token", nil
		},
		errOut: &output.Writer{Out: &stderr},
	})
	if !failed || len(entries) != 2 || entries[0].LatestVersion != "unknown" || entries[1].LatestVersion != "v2.0.0" {
		t.Fatalf("entries=%+v failed=%v", entries, failed)
	}
	if _, ok := c.Entries["broken"]; ok {
		t.Fatal("failed auth must not be cached")
	}
	if !bytes.Contains(stderr.Bytes(), []byte(`"broken" (GitHub account "missing")`)) {
		t.Fatalf("missing app/account diagnostic: %s", &stderr)
	}
	var stdout bytes.Buffer
	if err := (&output.Writer{Out: &stdout}).PrintJSON(entries); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("invalid JSON: %s", &stdout)
	}
}

func TestCheckRejectsAccountForNonGitHubBeforeCache(t *testing.T) {
	c := &cache.Cache{Entries: map[string]cache.Entry{}}
	cache.SetForApp(c, "tool", "v1.0.0", "v2.0.0", "work", false)
	entries, failed := collectCheckEntries([]config.App{{Name: "tool", GitHubUser: "work"}}, c, false, checkDependencies{
		runner: &stubRunner{infos: map[string]*goversion.Info{"tool": {Path: "example.org/tool", Version: "v1.0.0"}}},
		githubForApp: func(config.App) (github.Client, string, error) {
			t.Fatal("must not look up credentials")
			return nil, "", nil
		},
		errOut: &output.Writer{Out: &bytes.Buffer{}},
	})
	if !failed || entries[0].LatestVersion != "unknown" {
		t.Fatalf("entries=%+v failed=%v", entries, failed)
	}
}
