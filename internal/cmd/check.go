package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/gomodule"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type checkEntry struct {
	Name             string `json:"name"`
	InstalledVersion string `json:"installed_version"`
	LatestVersion    string `json:"latest_version"`
	UpdateAvailable  bool   `json:"update_available"`
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output as JSON")
	forceFlag := fs.Bool("force", false, "Refresh version information, ignoring cache")
	_ = fs.Parse(args)

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

	entries, failed := collectCheckEntries(cfg.Apps, c, *forceFlag, checkDependencies{
		runner:       &goversion.DefaultRunner{},
		githubForApp: newAppGitHubResolver(cfg),
		resolver:     gomodule.NewDefaultResolverWithGOPROXY(cfg.GOPROXY),
		errOut:       output.ErrorWriter,
	})

	// Save updated cache
	_ = cache.Save(cachePath, c)

	if *jsonFlag {
		if err := output.PrintJSON(entries); err != nil {
			output.Error(fmt.Sprintf("Failed to output JSON: %v", err))
			os.Exit(1)
		}
		if failed {
			os.Exit(1)
		}
		return
	}

	// Calculate column widths
	nameW := len("Name")
	instW := len("Installed")
	latW := len("Latest")
	updW := len("Update")
	for _, e := range entries {
		if len(e.Name) > nameW {
			nameW = len(e.Name)
		}
		if len(e.InstalledVersion) > instW {
			instW = len(e.InstalledVersion)
		}
		if len(e.LatestVersion) > latW {
			latW = len(e.LatestVersion)
		}
	}

	output.Header("Update Check")
	fmt.Println()
	// Header row
	fmt.Printf("  %s%s%-*s  %-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "Name", instW, "Installed", latW, "Latest", updW, "Update", output.Reset)
	// Separator
	fmt.Printf("  %s%s  %s  %s  %s%s\n", output.Gray,
		strings.Repeat("─", nameW), strings.Repeat("─", instW), strings.Repeat("─", latW), strings.Repeat("─", updW), output.Reset)
	// Data rows
	for _, e := range entries {
		updateStr := "no"
		updateColor := output.Gray
		if e.UpdateAvailable {
			updateStr = "yes"
			updateColor = output.Yellow
		}
		fmt.Printf("  %-*s  %s%-*s%s  %s%-*s%s  %s%-*s%s\n",
			nameW, e.Name,
			output.Green, instW, e.InstalledVersion, output.Reset,
			output.Cyan, latW, e.LatestVersion, output.Reset,
			updateColor, updW, updateStr, output.Reset)
	}
	fmt.Println()
	if failed {
		os.Exit(1)
	}
}

type checkDependencies struct {
	runner       goversion.Runner
	githubForApp appGitHubResolver
	resolver     gomodule.Resolver
	errOut       *output.Writer
}

func collectCheckEntries(apps []config.App, c *cache.Cache, force bool, deps checkDependencies) ([]checkEntry, bool) {
	entries := make([]checkEntry, 0, len(apps))
	failed := false
	for _, app := range apps {
		entry := checkEntry{Name: app.Name, InstalledVersion: "unknown", LatestVersion: "unknown"}
		info, err := deps.runner.GetInfo(app.Name)
		if err != nil {
			deps.errOut.Warn(fmt.Sprintf("Could not get info for %s: %v", appDescription(app), err))
			failed = true
			entries = append(entries, entry)
			continue
		}
		entry.InstalledVersion = info.Version
		if err := validateAppGitHub(app, info.Path); err != nil {
			deps.errOut.Error(fmt.Sprintf("Cannot check %s: %v", appDescription(app), err))
			failed = true
			entries = append(entries, entry)
			continue
		}
		cached, found := cache.Get(c, app.Name)
		if !force && found && cached.Matches(info.Version, app.GitHubUser, app.Private) && !cache.IsExpired(cached, cache.DefaultTTL) {
			entry.LatestVersion = cached.LatestVersion
			entry.UpdateAvailable = entry.InstalledVersion != entry.LatestVersion
		} else {
			var ghClient github.Client
			if goversion.IsGitHubRepo(info.Path) {
				ghClient, _, err = deps.githubForApp(app)
				if err != nil {
					deps.errOut.Error(fmt.Sprintf("Cannot authenticate %s: %v", appDescription(app), err))
					failed = true
					entries = append(entries, entry)
					continue
				}
			}
			result, err := checkForUpdate(info.Path, info.Version, ghClient, deps.resolver)
			if err != nil {
				deps.errOut.Warn(fmt.Sprintf("Could not fetch latest version for %s: %v", appDescription(app), err))
				failed = true
				entries = append(entries, entry)
				continue
			}
			entry.LatestVersion = result.latestVersion
			entry.UpdateAvailable = result.updateAvailable
			cache.SetForApp(c, app.Name, info.Version, result.latestVersion, app.GitHubUser, app.Private)
		}
		entries = append(entries, entry)
	}
	return entries, failed
}
