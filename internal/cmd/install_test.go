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
