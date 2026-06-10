package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func TestParseOwnerRepo(t *testing.T) {
	tests := []struct {
		input     string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"owner/repo", "owner", "repo", false},
		{"github.com/owner/repo", "owner", "repo", false},
		{"UnitVectorY-Labs/gogitup", "UnitVectorY-Labs", "gogitup", false},
		{"github.com/UnitVectorY-Labs/gogitup", "UnitVectorY-Labs", "gogitup", false},
		{"invalid", "", "", true},
		{"", "", "", true},
		{"/repo", "", "", true},
		{"owner/", "", "", true},
	}

	for _, tc := range tests {
		owner, repo, err := parseOwnerRepo(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseOwnerRepo(%q) expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseOwnerRepo(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if owner != tc.wantOwner || repo != tc.wantRepo {
			t.Errorf("parseOwnerRepo(%q) = (%q, %q), want (%q, %q)", tc.input, owner, repo, tc.wantOwner, tc.wantRepo)
		}
	}
}

func TestParseInstallTarget(t *testing.T) {
	tests := []struct {
		input       string
		packagePath string
		owner       string
		repo        string
		wantErr     bool
	}{
		{"owner/repo", "github.com/owner/repo", "owner", "repo", false},
		{"github.com/owner/repo", "github.com/owner/repo", "owner", "repo", false},
		{"golang.org/x/vuln/cmd/govulncheck", "golang.org/x/vuln/cmd/govulncheck", "", "", false},
		{"golang.org/x/vuln/cmd/govulncheck@latest", "golang.org/x/vuln/cmd/govulncheck", "", "", false},
		{"github.com/owner/repo/cmd/tool", "github.com/owner/repo/cmd/tool", "owner", "repo", false},
		{"owner/repo@v1.2.0", "", "", "", true},
		{"golang.org/x/vuln/cmd/govulncheck@v1.2.0", "", "", "", true},
		{"invalid", "", "", "", true},
		{"owner/repo/extra", "", "", "", true},
	}

	for _, tc := range tests {
		target, err := parseInstallTarget(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseInstallTarget(%q) expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseInstallTarget(%q) unexpected error: %v", tc.input, err)
			continue
		}
		if target.packagePath != tc.packagePath || target.owner != tc.owner || target.repo != tc.repo {
			t.Errorf("parseInstallTarget(%q) = %#v, want package=%q owner=%q repo=%q",
				tc.input, target, tc.packagePath, tc.owner, tc.repo)
		}
	}
}

func TestInstallTargetInstallPath(t *testing.T) {
	tests := []struct {
		name   string
		target installTarget
		want   string
	}{
		{
			name: "GitHub module root",
			target: installTarget{
				packagePath: "github.com/owner/repo",
				owner:       "owner",
				repo:        "repo",
			},
			want: "",
		},
		{
			name: "GitHub command package",
			target: installTarget{
				packagePath: "github.com/owner/repo/cmd/tool",
				owner:       "owner",
				repo:        "repo",
			},
			want: "github.com/owner/repo/cmd/tool",
		},
		{
			name: "non-GitHub command package",
			target: installTarget{
				packagePath: "golang.org/x/vuln/cmd/govulncheck",
			},
			want: "golang.org/x/vuln/cmd/govulncheck",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.target.installPath(); got != tc.want {
				t.Fatalf("installPath() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRunInstallAppSuccess(t *testing.T) {
	ghClient := &stubGitHubClient{
		releases: map[string]string{
			"acme/mytool": "v2.0.0",
		},
	}
	inst := &stubInstaller{}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"mytool": {Path: "github.com/acme/mytool", Version: "v2.0.0"},
		},
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	binaryName, err := runInstallApp("acme", "mytool", installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if binaryName != "mytool" {
		t.Fatalf("expected binaryName 'mytool', got %q", binaryName)
	}
	if len(inst.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(inst.calls))
	}
	if inst.calls[0].modulePath != "github.com/acme/mytool" || inst.calls[0].version != "v2.0.0" {
		t.Fatalf("unexpected install call: %+v", inst.calls[0])
	}
	if !strings.Contains(stdout.String(), "github.com/acme/mytool@v2.0.0") {
		t.Fatalf("expected output to mention module path, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr, got %q", stderr.String())
	}
}

func TestRunInstallTargetNonGitHub(t *testing.T) {
	ghClient := &stubGitHubClient{}
	inst := &stubInstaller{}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"govulncheck": {
				Path:        "golang.org/x/vuln",
				PackagePath: "golang.org/x/vuln/cmd/govulncheck",
				Version:     "v1.2.3",
			},
		},
	}
	var stdout, stderr bytes.Buffer

	binaryName, err := runInstallTarget(installTarget{
		packagePath: "golang.org/x/vuln/cmd/govulncheck",
	}, installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if binaryName != "govulncheck" {
		t.Fatalf("expected govulncheck, got %q", binaryName)
	}
	if len(inst.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(inst.calls))
	}
	if inst.calls[0].modulePath != "golang.org/x/vuln/cmd/govulncheck" || inst.calls[0].version != "latest" {
		t.Fatalf("unexpected install call: %+v", inst.calls[0])
	}
}

func TestRunInstallTargetGitHubSubpathUsesRelease(t *testing.T) {
	ghClient := &stubGitHubClient{
		releases: map[string]string{"owner/repo": "v1.2.0"},
	}
	inst := &stubInstaller{}
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"tool": {
				Path:        "github.com/owner/repo",
				PackagePath: "github.com/owner/repo/cmd/tool",
				Version:     "v1.2.0",
			},
		},
	}
	target, err := parseInstallTarget("github.com/owner/repo/cmd/tool")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	binaryName, err := runInstallTarget(target, installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &bytes.Buffer{}},
		errOut:    &output.Writer{Out: &bytes.Buffer{}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if binaryName != "tool" {
		t.Fatalf("expected tool, got %q", binaryName)
	}
	if target.installPath() != "github.com/owner/repo/cmd/tool" {
		t.Fatalf("unexpected install path %q", target.installPath())
	}
	if len(inst.calls) != 1 {
		t.Fatalf("expected 1 install call, got %d", len(inst.calls))
	}
	if inst.calls[0].modulePath != "github.com/owner/repo/cmd/tool" || inst.calls[0].version != "v1.2.0" {
		t.Fatalf("unexpected install call: %+v", inst.calls[0])
	}
}

func TestRunInstallAppGitHubError(t *testing.T) {
	ghClient := &stubGitHubClient{
		errs: map[string]error{
			"acme/mytool": errors.New("not found"),
		},
	}
	inst := &stubInstaller{}
	runner := &stubRunner{}

	var stdout, stderr bytes.Buffer

	_, err := runInstallApp("acme", "mytool", installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if err == nil {
		t.Fatal("expected error when GitHub client fails, got nil")
	}
	if len(inst.calls) != 0 {
		t.Fatalf("expected no install calls, got %d", len(inst.calls))
	}
}

func TestRunInstallAppInstallError(t *testing.T) {
	ghClient := &stubGitHubClient{
		releases: map[string]string{"acme/mytool": "v1.0.0"},
	}
	inst := &stubInstaller{err: errors.New("exit status 1")}
	runner := &stubRunner{}

	var stdout, stderr bytes.Buffer

	_, err := runInstallApp("acme", "mytool", installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if err == nil {
		t.Fatal("expected error when install fails, got nil")
	}
}

func TestRunInstallAppBinaryNotOnPath(t *testing.T) {
	ghClient := &stubGitHubClient{
		releases: map[string]string{"acme/mytool": "v1.0.0"},
	}
	inst := &stubInstaller{}
	runner := &stubRunner{
		errs: map[string]error{
			"mytool": errors.New("binary not found"),
		},
	}

	var stdout, stderr bytes.Buffer

	_, err := runInstallApp("acme", "mytool", installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       &output.Writer{Out: &stdout},
		errOut:    &output.Writer{Out: &stderr},
	})

	if err == nil {
		t.Fatal("expected error when binary not found on PATH, got nil")
	}
	if !strings.Contains(err.Error(), "not found on PATH") {
		t.Fatalf("expected 'not found on PATH' in error, got %q", err.Error())
	}
}
