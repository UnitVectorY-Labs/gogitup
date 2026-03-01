package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/installer"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type updateOptions struct {
	Verbose bool
}

type updateDependencies struct {
	runner    goversion.Runner
	ghClient  github.Client
	installer installer.Installer
	out       *output.Writer
	errOut    *output.Writer
}

func parseUpdateOptions(args []string, stderr io.Writer) (updateOptions, error) {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	fs.SetOutput(stderr)

	verboseFlag := fs.Bool("verbose", false, "Show binaries that are already up to date")
	if err := fs.Parse(args); err != nil {
		return updateOptions{}, err
	}

	return updateOptions{Verbose: *verboseFlag}, nil
}

func runUpdate(args []string) {
	opts, err := parseUpdateOptions(args, output.ErrorWriter.Out)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		os.Exit(2)
	}

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
	deps := updateDependencies{
		runner:    runner,
		ghClient:  ghClient,
		installer: inst,
		out:       output.DefaultWriter,
		errOut:    output.ErrorWriter,
	}
	updated := runUpdateApps(cfg, c, opts, deps)

	// Save updated cache
	_ = cache.Save(cachePath, c)

	fmt.Println()
	if updated == 0 {
		deps.out.Info("All binaries are up to date.")
	} else {
		deps.out.Success(fmt.Sprintf("Updated %d binary(ies).", updated))
	}
}

func runUpdateApps(cfg *config.Config, c *cache.Cache, opts updateOptions, deps updateDependencies) int {
	updated := 0

	for _, app := range cfg.Apps {
		info, err := deps.runner.GetInfo(app.Name)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			continue
		}

		owner, repo, err := goversion.ParseGitHubRepo(info.Path)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not parse GitHub repo for '%s': %v", app.Name, err))
			continue
		}

		// Always re-check GitHub for latest version (ignore cache)
		latest, err := deps.ghClient.GetLatestRelease(owner, repo)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not fetch latest release for '%s': %v", app.Name, err))
			continue
		}

		cache.Set(c, app.Name, latest)

		if info.Version == latest {
			if opts.Verbose {
				deps.out.Info(updateUpToDateMessage(app.Name, info.Version))
			}
			continue
		}

		deps.out.StartProgress(updateProgressMessage(app.Name, info.Version, latest))

		_, err = deps.installer.Install(info.Path, latest)
		if err != nil {
			deps.errOut.Error(fmt.Sprintf("Failed to update '%s': %v", app.Name, err))
			continue
		}

		deps.out.Success(updateSuccessMessage(app.Name, latest))
		updated++
	}

	return updated
}

func updateUpToDateMessage(name, version string) string {
	return fmt.Sprintf("'%s' is already up to date (%s)", name, installedVersion(version))
}

func updateProgressMessage(name, currentVersion, latestVersion string) string {
	return fmt.Sprintf("Updating '%s' from %s to %s", name, installedVersion(currentVersion), latestVersionLabel(latestVersion))
}

func updateSuccessMessage(name, version string) string {
	return fmt.Sprintf("Updated '%s' to %s", name, installedVersion(version))
}

func installedVersion(version string) string {
	return output.Green + version + output.Reset
}

func latestVersionLabel(version string) string {
	return output.Cyan + version + output.Reset
}
