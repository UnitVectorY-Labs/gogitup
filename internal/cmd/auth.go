package cmd

import (
	"fmt"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
)

type appGitHubResolver func(config.App) (github.Client, string, error)

func newAppGitHubResolver(cfg *config.Config) appGitHubResolver {
	return func(app config.App) (github.Client, string, error) {
		token, err := github.ResolveAppToken(cfg.GitHubAuth, app.Private, app.GitHubUser)
		if err != nil {
			return nil, "", err
		}
		return github.NewDefaultClient(token), token, nil
	}
}

// Validate locally, including before a cache hit, without looking up credentials.
func validateAppGitHub(app config.App, modulePath string) error {
	if err := config.ValidateGitHubUser(app.GitHubUser); err != nil {
		return err
	}
	if app.GitHubUser != "" || app.Private {
		if _, _, err := goversion.ParseGitHubRepo(modulePath); err != nil {
			return fmt.Errorf("github_user and private are only supported for github.com repositories: %w", err)
		}
	}
	return nil
}

func appDescription(app config.App) string {
	if app.GitHubUser != "" {
		return fmt.Sprintf("%q (GitHub account %q)", app.Name, app.GitHubUser)
	}
	return fmt.Sprintf("%q", app.Name)
}
