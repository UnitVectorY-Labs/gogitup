package cmd

import (
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/installer"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func runUpdate(args []string) {
	_ = args

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	if len(cfg.Apps) == 0 {
		output.Info("No binaries registered. Use 'gogitup add <name>' to add one.")
		return
	}

	cachePath := cache.DefaultPath()
	c, err := cache.Load(cachePath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load cache: %v", err))
		os.Exit(1)
	}

	runner := &goversion.DefaultRunner{}
	ghClient := github.NewDefaultClient(github.ResolveToken(cfg.GitHubAuth))
	inst := installer.NewDefaultInstaller()

	updated := 0

	for _, app := range cfg.Apps {
		info, err := runner.GetInfo(app.Name)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			continue
		}

		owner, repo, err := goversion.ParseGitHubRepo(info.Path)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not parse GitHub repo for '%s': %v", app.Name, err))
			continue
		}

		// Always re-check GitHub for latest version (ignore cache)
		latest, err := ghClient.GetLatestRelease(owner, repo)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not fetch latest release for '%s': %v", app.Name, err))
			continue
		}

		cache.Set(c, app.Name, latest)

		if info.Version == latest {
			output.Info(fmt.Sprintf("'%s' is already up to date (%s)", app.Name, info.Version))
			continue
		}

		output.StartProgress(fmt.Sprintf("Updating '%s' from %s to %s", app.Name, info.Version, latest))

		_, err = inst.Install(info.Path, latest)
		if err != nil {
			output.Error(fmt.Sprintf("Failed to update '%s': %v", app.Name, err))
			continue
		}

		output.Success(fmt.Sprintf("Updated '%s' to %s", app.Name, latest))
		updated++
	}

	// Save updated cache
	_ = cache.Save(cachePath, c)

	fmt.Println()
	if updated == 0 {
		output.Info("All binaries are up to date.")
	} else {
		output.Success(fmt.Sprintf("Updated %d binary(ies).", updated))
	}
}
