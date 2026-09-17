package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ResolveAppToken selects credentials for one app. An explicit user always
// requires stored github.com credentials, without changing gh's active account.
func ResolveAppToken(useAuth, private bool, user string) (string, error) {
	if user != "" {
		return accountToken(user, os.Getenv, func(args ...string) ([]byte, error) {
			return exec.Command("gh", args...).Output()
		})
	}
	token := ResolveToken(useAuth || private)
	if private && token == "" {
		return "", fmt.Errorf("private GitHub access requires authentication; set GITHUB_TOKEN or run 'gh auth login'")
	}
	return token, nil
}

func accountToken(user string, getenv func(string) string, run func(...string) ([]byte, error)) (string, error) {
	if getenv("GH_TOKEN") != "" || getenv("GITHUB_TOKEN") != "" {
		return "", fmt.Errorf("github_user %q requires stored GitHub CLI credentials; unset GH_TOKEN and GITHUB_TOKEN", user)
	}
	data, err := run("auth", "token", "--hostname", "github.com", "--user", user)
	if err != nil {
		// Never include command output: it may contain credentials.
		return "", fmt.Errorf("cannot read GitHub CLI token for %q: ensure gh supports 'auth token --user' and log in with 'gh auth login --hostname github.com': %w", user, err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("GitHub CLI returned an empty token for %q; run 'gh auth login --hostname github.com'", user)
	}
	return token, nil
}
