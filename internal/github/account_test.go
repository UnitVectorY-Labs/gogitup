package github

import (
	"errors"
	"io"
	"net/http"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestAccountToken(t *testing.T) {
	for _, tt := range []struct {
		name, output, want string
		err                error
	}{
		{name: "stored account", output: "  work-token\n", want: "work-token"},
		{name: "empty token", output: " \n"},
		{name: "missing gh", err: exec.ErrNotFound},
		{name: "unknown account or unsupported flag", output: "secret-output", err: &exec.ExitError{Stderr: []byte("secret-stderr")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			token, err := accountToken("work", func(string) string { return "" }, func(args ...string) ([]byte, error) {
				calls++
				want := []string{"auth", "token", "--hostname", "github.com", "--user", "work"}
				if !reflect.DeepEqual(args, want) {
					t.Fatalf("args = %v, want %v", args, want)
				}
				return []byte(tt.output), tt.err
			})
			if calls != 1 || token != tt.want || (err != nil) != (tt.want == "") {
				t.Fatalf("calls=%d token=%q err=%v", calls, token, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "work") || strings.Contains(err.Error(), "secret")) {
				t.Fatalf("unsafe or unhelpful error: %v", err)
			}
		})
	}
}

func TestAccountTokenRejectsEnvironmentOverrides(t *testing.T) {
	for _, key := range []string{"GH_TOKEN", "GITHUB_TOKEN"} {
		t.Run(key, func(t *testing.T) {
			token, err := accountToken("work", func(name string) string {
				if name == key {
					return "env-secret"
				}
				return ""
			}, func(...string) ([]byte, error) {
				t.Fatal("must reject conflicting credentials before executing gh")
				return nil, nil
			})
			if token != "" || err == nil || !strings.Contains(err.Error(), "unset GH_TOKEN and GITHUB_TOKEN") {
				t.Fatalf("token=%q err=%v", token, err)
			}
		})
	}
}

func TestResolveAppTokenDefaultsAndRequiredCredentials(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "default-token")
	t.Setenv("GH_TOKEN", "")
	for _, tt := range []struct {
		auth, private bool
		want          string
	}{
		{false, false, ""}, {true, false, "default-token"}, {false, true, "default-token"},
	} {
		token, err := ResolveAppToken(tt.auth, tt.private, "")
		if token != tt.want || err != nil {
			t.Fatalf("options=%+v token=%q err=%v", tt, token, err)
		}
	}
	if token, err := ResolveAppToken(false, false, "work"); token != "" || err == nil {
		t.Fatal("explicit account must not fall back to environment")
	}
	// An empty PATH guarantees these tests cannot execute a real gh command.
	t.Setenv("PATH", t.TempDir())
	t.Setenv("GITHUB_TOKEN", "")
	if token, err := ResolveAppToken(true, false, ""); token != "" || err != nil {
		t.Fatalf("public fallback: %q %v", token, err)
	}
	if _, err := ResolveAppToken(false, true, ""); err == nil {
		t.Fatal("private app must require credentials")
	}
	if _, err := ResolveAppToken(false, false, "work"); err == nil {
		t.Fatal("selected account must require gh")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSelectedTokenAuthenticatesReleaseRequest(t *testing.T) {
	token, err := accountToken("work", func(string) string { return "" }, func(...string) ([]byte, error) { return []byte("work-token\n"), nil })
	if err != nil {
		t.Fatal(err)
	}
	client := NewDefaultClient(token)
	client.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.github.com/repos/other-owner/tool/releases/latest" || req.Header.Get("Authorization") != "Bearer work-token" {
			return nil, errors.New("wrong repository or credentials")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"tag_name":"v2.0.0"}`)), Header: make(http.Header)}, nil
	})
	version, err := client.GetLatestRelease("other-owner", "tool")
	if err != nil || version != "v2.0.0" {
		t.Fatalf("version=%q err=%v", version, err)
	}
}
