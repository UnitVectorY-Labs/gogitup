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
	GoVersion        string `json:"go_version"`
	GoVersionNewer   bool   `json:"go_version_newer"`
	goVersionRaw     string
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

	runner := &goversion.DefaultRunner{}
	githubToken := github.ResolveToken(cfg.GitHubAuth || config.HasPrivateApps(cfg))
	if config.HasPrivateApps(cfg) && githubToken == "" {
		output.Error("Checking private GitHub repositories requires authentication; set GITHUB_TOKEN or run 'gh auth login'")
		os.Exit(1)
	}
	ghClient := github.NewDefaultClient(githubToken)
	moduleResolver := gomodule.NewDefaultResolverWithGOPROXY(cfg.GOPROXY)

	activeGoVersion, err := goversion.CurrentToolchainVersion()
	if err != nil {
		activeGoVersion = ""
	}

	entries := make([]checkEntry, 0, len(cfg.Apps))

	for _, app := range cfg.Apps {
		entry := checkEntry{Name: app.Name, InstalledVersion: "unknown", LatestVersion: "unknown", GoVersion: "unknown"}

		info, err := runner.GetInfo(app.Name)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			entries = append(entries, entry)
			continue
		}
		entry.InstalledVersion = info.Version
		entry.goVersionRaw = info.GoVersion
		entry.GoVersion = strings.TrimPrefix(info.GoVersion, "go")
		if info.GoVersion == "" {
			entry.GoVersion = "unknown"
		}
		if activeGoVersion != "" && info.GoVersion != "" {
			if newer, nerr := goversion.IsToolchainNewer(activeGoVersion, info.GoVersion); nerr == nil {
				entry.GoVersionNewer = newer
			}
		}

		// Cached update decisions are valid only for the installed version checked.
		cached, found := cache.Get(c, app.Name)
		if !*forceFlag && found && cached.InstalledVersion == info.Version && !cache.IsExpired(cached, cache.DefaultTTL) {
			entry.LatestVersion = cached.LatestVersion
			entry.UpdateAvailable = entry.InstalledVersion != entry.LatestVersion
		} else {
			result, err := checkForUpdate(info.Path, info.Version, ghClient, moduleResolver)
			if err != nil {
				output.Warn(fmt.Sprintf("Could not fetch latest version for '%s': %v", app.Name, err))
				entries = append(entries, entry)
				continue
			}
			entry.LatestVersion = result.latestVersion
			entry.UpdateAvailable = result.updateAvailable
			cache.SetForInstalledVersion(c, app.Name, info.Version, result.latestVersion)
		}

		entries = append(entries, entry)
	}

	// Save updated cache
	_ = cache.Save(cachePath, c)

	if *jsonFlag {
		if err := output.PrintJSON(entries); err != nil {
			output.Error(fmt.Sprintf("Failed to output JSON: %v", err))
			os.Exit(1)
		}
		return
	}

	// Calculate column widths
	nameW := len("Name")
	instW := len("Installed")
	latW := len("Latest")
	updW := len("Update")
	const goHeader = "Version"
	goW := len(goHeader)
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
		if len(e.GoVersion) > goW {
			goW = len(e.GoVersion)
		}
	}

	output.Header("Update Check")
	fmt.Println()
	// Header rows (two-line style to match list)
	fmt.Printf("  %s%s%-*s  %-*s  %-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "", instW, "", latW, "", updW, "", goW, "Go", output.Reset)
	fmt.Printf("  %s%s%-*s  %-*s  %-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "Name", instW, "Installed", latW, "Latest", updW, "Update", goW, goHeader, output.Reset)
	// Separator
	fmt.Printf("  %s%s  %s  %s  %s  %s%s\n", output.Gray,
		strings.Repeat("─", nameW), strings.Repeat("─", instW), strings.Repeat("─", latW), strings.Repeat("─", updW), strings.Repeat("─", goW), output.Reset)
	// Data rows
	for _, e := range entries {
		updateStr := "no"
		updateColor := output.Gray
		if e.UpdateAvailable {
			updateStr = "yes"
			updateColor = output.Yellow
		}
		goColor := checkGoVersionColor(e)
		fmt.Printf("  %-*s  %s%-*s%s  %s%-*s%s  %s%-*s%s  %s%-*s%s\n",
			nameW, e.Name,
			output.Green, instW, e.InstalledVersion, output.Reset,
			output.Cyan, latW, e.LatestVersion, output.Reset,
			updateColor, updW, updateStr, output.Reset,
			goColor, goW, e.GoVersion, output.Reset)
	}
	fmt.Println()
}

func checkGoVersionColor(e checkEntry) string {
	if e.goVersionRaw == "" {
		return output.Gray
	}
	if e.GoVersionNewer {
		return output.Yellow
	}
	return output.Green
}
