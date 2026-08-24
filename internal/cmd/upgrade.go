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
	Verbose            bool
	RebuildWithNewerGo bool
	DryRun             bool
}

type upgradeDependencies struct {
	runner      goversion.Runner
	ghClient    github.Client
	resolver    gomodule.Resolver
	installer   installer.Installer
	githubToken string
	out         *output.Writer
	errOut      *output.Writer
	currentGo   func() (string, error)
}

type updateResult struct {
	latestVersion   string
	updateAvailable bool
}

func parseUpgradeOptions(args []string, stderr io.Writer) (upgradeOptions, error) {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(stderr)

	verboseFlag := fs.Bool("verbose", false, "Show binaries that are already up to date")
	rebuildFlag := fs.Bool("go-version", false, "Rebuild binaries compiled with an older Go toolchain")
	dryRunFlag := fs.Bool("dry-run", false, "List updates without installing them")
	if err := fs.Parse(args); err != nil {
		return upgradeOptions{}, err
	}

	return upgradeOptions{Verbose: *verboseFlag, RebuildWithNewerGo: *rebuildFlag, DryRun: *dryRunFlag}, nil
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
	githubToken := github.ResolveToken(cfg.GitHubAuth || config.HasPrivateApps(cfg))
	if config.HasPrivateApps(cfg) && githubToken == "" {
		output.Error("Upgrading private GitHub repositories requires authentication; set GITHUB_TOKEN or run 'gh auth login'")
		os.Exit(1)
	}
	ghClient := github.NewDefaultClient(githubToken)
	inst := installer.NewDefaultInstallerWithOptions(cfg.GOPROXY, cfg.CGOEnabled)
	deps := upgradeDependencies{
		runner:      runner,
		ghClient:    ghClient,
		resolver:    gomodule.NewDefaultResolverWithGOPROXY(cfg.GOPROXY),
		installer:   inst,
		githubToken: githubToken,
		out:         output.DefaultWriter,
		errOut:      output.ErrorWriter,
		currentGo:   goversion.CurrentToolchainVersion,
	}
	updated := runUpgradeApps(cfg, c, opts, deps)

	if !opts.DryRun {
		// Save updated cache only when performing the requested updates.
		_ = cache.Save(cachePath, c)
	}

	fmt.Println()
	if opts.DryRun {
		if updated == 0 {
			deps.out.Info("Dry run: no binaries would be updated.")
		} else {
			deps.out.Info(fmt.Sprintf("Dry run: %d binary(ies) would be updated.", updated))
		}
		return
	}
	if updated == 0 {
		deps.out.Info("All binaries are up to date.")
	} else {
		deps.out.Success(fmt.Sprintf("Upgraded %d binary(ies).", updated))
	}
}

func runUpgradeApps(cfg *config.Config, c *cache.Cache, opts upgradeOptions, deps upgradeDependencies) int {
	updated := 0
	activeGoVersion := ""
	if opts.RebuildWithNewerGo {
		if deps.currentGo == nil {
			deps.currentGo = goversion.CurrentToolchainVersion
		}
		var err error
		activeGoVersion, err = deps.currentGo()
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not determine active Go toolchain version: %v", err))
			activeGoVersion = ""
		}
	}

	for _, app := range cfg.Apps {
		info, err := deps.runner.GetInfo(app.Name)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			continue
		}

		rebuild := false
		if activeGoVersion != "" {
			rebuild, err = goversion.IsToolchainNewer(activeGoVersion, info.GoVersion)
			if err != nil {
				deps.out.Warn(fmt.Sprintf("Could not compare Go toolchain version for '%s': %v", app.Name, err))
			}
		}

		// Always perform a fresh update check (ignore cache).
		result, err := checkForUpdate(info.Path, info.Version, deps.ghClient, deps.resolver)
		if err != nil {
			deps.out.Warn(fmt.Sprintf("Could not fetch latest version for '%s': %v", app.Name, err))
			if !rebuild {
				continue
			}
			result = updateResult{latestVersion: info.Version}
		} else if !opts.DryRun {
			cache.SetForInstalledVersion(c, app.Name, info.Version, result.latestVersion)
		}

		if !result.updateAvailable && !rebuild {
			if opts.Verbose {
				deps.out.Info(upgradeUpToDateMessage(app.Name, info.Version))
			}
			continue
		}

		installVersion := result.latestVersion
		if rebuild && !result.updateAvailable {
			installVersion = info.Version
			if opts.DryRun {
				deps.out.Info(dryRunRebuildMessage(app.Name, installVersion, info.GoVersion, activeGoVersion))
				updated++
				continue
			}
			deps.out.StartProgress(rebuildProgressMessage(app.Name, info.Version, info.GoVersion, activeGoVersion))
		} else {
			if opts.DryRun {
				deps.out.Info(dryRunUpgradeMessage(app.Name, info.Version, result.latestVersion))
				updated++
				continue
			}
			deps.out.StartProgress(upgradeProgressMessage(app.Name, info.Version, result.latestVersion))
		}

		installPath := app.InstallPath
		if installPath == "" {
			installPath = info.PackagePath
		}
		if installPath == "" {
			installPath = info.Path
		}
		installerOptions := installer.InstallOptions{}
		if app.Private {
			installerOptions = installer.InstallOptions{
				PrivateModule: info.Path,
				GitHubToken:   deps.githubToken,
			}
		}
		_, err = deps.installer.Install(installPath, installVersion, installerOptions)
		if err != nil {
			deps.errOut.Error(fmt.Sprintf("Failed to update '%s': %v", app.Name, err))
			continue
		}

		if rebuild && !result.updateAvailable {
			deps.out.Success(rebuildSuccessMessage(app.Name, installVersion, activeGoVersion))
		} else {
			deps.out.Success(upgradeSuccessMessage(app.Name, installVersion))
		}
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

func rebuildProgressMessage(name, version, oldGoVersion, newGoVersion string) string {
	return fmt.Sprintf("Rebuilding '%s' %s with %s (was %s)", name, installedVersion(version), newGoVersion, oldGoVersion)
}

func rebuildSuccessMessage(name, version, goVersion string) string {
	return fmt.Sprintf("Rebuilt '%s' %s with %s", name, installedVersion(version), goVersion)
}

func dryRunUpgradeMessage(name, currentVersion, latestVersion string) string {
	return fmt.Sprintf("Would upgrade '%s' from %s to %s", name, installedVersion(currentVersion), latestVersionLabel(latestVersion))
}

func dryRunRebuildMessage(name, version, oldGoVersion, newGoVersion string) string {
	return fmt.Sprintf("Would rebuild '%s' %s with %s (currently %s)", name, installedVersion(version), newGoVersion, oldGoVersion)
}

func installedVersion(version string) string {
	return output.Green + version + output.Reset
}

func latestVersionLabel(version string) string {
	return output.Cyan + version + output.Reset
}
