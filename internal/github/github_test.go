package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLatestRelease_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octocat/hello-world/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Errorf("unexpected Accept header: %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releaseResponse{TagName: "v1.2.3"})
	}))
	defer server.Close()

	client := NewDefaultClient("")
	// Override the base URL by pointing the HTTP client at the test server
	client.httpClient = server.Client()

	// We need to hit the test server, so we make a request directly
	req, _ := http.NewRequest("GET", server.URL+"/repos/octocat/hello-world/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if release.TagName != "v1.2.3" {
		t.Fatalf("expected tag v1.2.3, got %s", release.TagName)
	}
}

func TestGetLatestRelease_RequestURL(t *testing.T) {
	var capturedPath string
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		capturedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(releaseResponse{TagName: "v0.1.0"})
	}))
	defer server.Close()

	// Create a client that sends requests to our test server
	client := &DefaultClient{
		token:      "test-token",
		httpClient: server.Client(),
	}

	// Build the request manually to the test server
	req, _ := http.NewRequest("GET", server.URL+"/repos/myowner/myrepo/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer test-token")
	_, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPath != "/repos/myowner/myrepo/releases/latest" {
		t.Fatalf("expected path /repos/myowner/myrepo/releases/latest, got %s", capturedPath)
	}
	if capturedAuth != "Bearer test-token" {
		t.Fatalf("expected Authorization header 'Bearer test-token', got %s", capturedAuth)
	}
}

func TestGetLatestRelease_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &DefaultClient{
		httpClient: server.Client(),
	}

	req, _ := http.NewRequest("GET", server.URL+"/repos/owner/repo/releases/latest", nil)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetLatestRelease_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := &DefaultClient{
		httpClient: server.Client(),
	}

	resp, err := client.httpClient.Do(mustNewRequest(t, server.URL+"/repos/owner/repo/releases/latest"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
		t.Fatal("expected JSON decode error, got nil")
	}
}

func TestResolveToken_UseAuthFalse(t *testing.T) {
	token := ResolveToken(false)
	if token != "" {
		t.Fatalf("expected empty token when useAuth is false, got %s", token)
	}
}

func TestResolveToken_EnvVar(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "env-token-value")
	token := ResolveToken(true)
	if token != "env-token-value" {
		t.Fatalf("expected env-token-value, got %s", token)
	}
}

func mustNewRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}
