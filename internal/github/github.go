package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Client is an interface for retrieving the latest release version from GitHub.
type Client interface {
	GetLatestRelease(owner, repo string) (string, error)
}

// DefaultClient implements Client using the GitHub REST API.
type DefaultClient struct {
	token      string
	httpClient *http.Client
}

// releaseResponse represents the relevant fields from the GitHub releases API.
type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// NewDefaultClient creates a new DefaultClient with an optional auth token.
func NewDefaultClient(token string) *DefaultClient {
	return &DefaultClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetLatestRelease fetches the latest release tag name for the given owner/repo.
func (c *DefaultClient) GetLatestRelease(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d for %s/%s", resp.StatusCode, owner, repo)
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release response: %w", err)
	}

	if release.TagName == "" {
		return "", errors.New("no release tag found for " + owner + "/" + repo)
	}

	return release.TagName, nil
}

// ResolveToken determines the GitHub token to use for API requests.
// If useAuth is false, returns an empty string. Otherwise it checks the
// GITHUB_TOKEN environment variable and falls back to the gh CLI.
func ResolveToken(useAuth bool) string {
	if !useAuth {
		return ""
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
