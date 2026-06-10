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
	"github.com/UnitVectorY-Labs/gogitup/internal/gomodule"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/installer"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type upgradeOptions struct {
	Verbose bool
}

type upgradeDependencies struct {
	runner    goversion.Runner
	ghClient  github.Client
	resolver  gomodule.Resolver
	installer installer.Installer
	out       *output.Writer
	errOut    *output.Writer
}

type updateResult struct {
	latestVersion   string
	updateAvailable bool
}

func parseUpgradeOptions(args []string, stderr io.Writer) (upgradeOptions, error) {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(stderr)

	verboseFlag := fs.Bool("verbose", false, "Show binaries that are already up to date")
	if err := fs.Parse(args); err != nil {
		return upgradeOptions{}, err
	}

	return upgradeOptions{Verbose: *verboseFlag}, nil
}

func runUpgrade(args []string) {
	opts, err := parseUpgradeOptions(args, output.ErrorWriter.Out)
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
	inst := installer.NewDefaultInstallerWithOptions(cfg.GOPROXY, cfg.CGOEnabled)
	deps := upgradeDependencies{
		runner:    runner,
		ghClient:  ghClient,
		resolver:  gomodule.NewDefaultResolverWithGOPROXY(cfg.GOPROXY),
		installer: inst,
		out:       output.DefaultWriter,
		errOut:    output.ErrorWriter,
	}
	updated := runUpgradeApps(cfg, c, opts, deps)

	// Save updated cache
	_ = cache.Save(cachePath, c)

	fmt.Println()
	if updated == 0 {
		deps.out.Info("All binaries are up to date.")
	} else {
		deps.out.Success(fmt.Sprintf("Upgraded %d binary(ies).", updated))
	}
}

func runUpgradeApps(cfg *config.Config, c *cache.Cache, opts upgradeOptions, deps upgradeDependencies) int {
	updated := 0

	for _, app := range cfg.Apps {
		info, err := deps.runner.GetInfo(app.Name)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			continue
		}

		// Always perform a fresh update check (ignore cache).
		result, err := checkForUpdate(info.Path, info.Version, deps.ghClient, deps.resolver)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not fetch latest version for '%s': %v", app.Name, err))
			continue
		}

		cache.SetForInstalledVersion(c, app.Name, info.Version, result.latestVersion)

		if !result.updateAvailable {
			if opts.Verbose {
				deps.out.Info(upgradeUpToDateMessage(app.Name, info.Version))
			}
			continue
		}

		deps.out.StartProgress(upgradeProgressMessage(app.Name, info.Version, result.latestVersion))

		installPath := app.InstallPath
		if installPath == "" {
			installPath = info.PackagePath
		}
		if installPath == "" {
			installPath = info.Path
		}
		_, err = deps.installer.Install(installPath, result.latestVersion)
		if err != nil {
			deps.errOut.Error(fmt.Sprintf("Failed to upgrade '%s': %v", app.Name, err))
			continue
		}

		deps.out.Success(upgradeSuccessMessage(app.Name, result.latestVersion))
		updated++
	}

	return updated
}

func checkForUpdate(modulePath, installedVersion string, ghClient github.Client, resolver gomodule.Resolver) (updateResult, error) {
	if goversion.IsGitHubRepo(modulePath) {
		owner, repo, err := goversion.ParseGitHubRepo(modulePath)
		if err != nil {
			return updateResult{}, err
		}
		latest, err := ghClient.GetLatestRelease(owner, repo)
		if err != nil {
			return updateResult{}, err
		}
		return updateResult{
			latestVersion:   latest,
			updateAvailable: installedVersion != latest,
		}, nil
	}
	result, err := resolver.Check(modulePath, installedVersion)
	if err != nil {
		return updateResult{}, err
	}
	return updateResult{
		latestVersion:   result.LatestVersion,
		updateAvailable: result.UpdateAvailable,
	}, nil
}

func upgradeUpToDateMessage(name, version string) string {
	return fmt.Sprintf("'%s' is already up to date (%s)", name, installedVersion(version))
}

func upgradeProgressMessage(name, currentVersion, latestVersion string) string {
	return fmt.Sprintf("Upgrading '%s' from %s to %s", name, installedVersion(currentVersion), latestVersionLabel(latestVersion))
}

func upgradeSuccessMessage(name, version string) string {
	return fmt.Sprintf("Upgraded '%s' to %s", name, installedVersion(version))
}

func installedVersion(version string) string {
	return output.Green + version + output.Reset
}

func latestVersionLabel(version string) string {
	return output.Cyan + version + output.Reset
}
