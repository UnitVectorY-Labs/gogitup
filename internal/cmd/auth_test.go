package cmd

import (
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
)

func TestAccountFlags(t *testing.T) {
	for _, args := range [][]string{{"--github-user", "work", "tool"}, {"--github-user=work", "tool"}} {
		opts, err := parseAddOptions(args)
		if err != nil || opts.GitHubUser != "work" || opts.Name != "tool" {
			t.Fatalf("add: %+v %v", opts, err)
		}
	}
	opts, err := parseInstallOptions([]string{"--private", "--github-user", "work", "acme/tool"})
	if err != nil || opts.GitHubUser != "work" || !opts.Private || opts.Target != "acme/tool" {
		t.Fatalf("install: %+v %v", opts, err)
	}
	for _, args := range [][]string{nil, {"--github-user"}, {"--github-user", "work"}, {"--github-user", " work ", "acme/tool"}, {"acme/tool", "--github-user", "work"}} {
		if _, err := parseAddOptions(args); err == nil {
			t.Errorf("add accepted invalid arguments %v", args)
		}
		if _, err := parseInstallOptions(args); err == nil {
			t.Errorf("install accepted invalid arguments %v", args)
		}
	}
}

func TestValidateAppGitHub(t *testing.T) {
	for _, tt := range []struct {
		app     config.App
		path    string
		wantErr bool
	}{
		{config.App{GitHubUser: "work"}, "github.com/other-owner/tool/v2", false},
		{config.App{GitHubUser: "work"}, "golang.org/x/tools", true},
		{config.App{GitHubUser: "work"}, "github.com/owner", true},
		{config.App{GitHubUser: "work\n"}, "github.com/owner/tool", true},
		{config.App{Private: true}, "golang.org/x/tools", true},
		{config.App{}, "golang.org/x/tools", false},
	} {
		if err := validateAppGitHub(tt.app, tt.path); (err != nil) != tt.wantErr {
			t.Errorf("%+v %q: %v", tt.app, tt.path, err)
		}
	}
}

func TestAppResolverKeepsDefaultAndNamedAccountsSeparate(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "default-token")
	t.Setenv("GH_TOKEN", "")
	resolve := newAppGitHubResolver(&config.Config{})
	if _, token, err := resolve(config.App{Private: true}); token != "default-token" || err != nil {
		t.Fatalf("private default: %q %v", token, err)
	}
	if client, token, err := resolve(config.App{GitHubUser: "work"}); client != nil || token != "" || err == nil {
		t.Fatal("named account must reject environment token")
	}
	if _, token, err := resolve(config.App{}); token != "" || err != nil {
		t.Fatalf("public default inherited other app's auth: %q %v", token, err)
	}
	resolve = newAppGitHubResolver(&config.Config{GitHubAuth: true})
	if _, token, err := resolve(config.App{}); token != "default-token" || err != nil {
		t.Fatalf("global auth: %q %v", token, err)
	}
}
