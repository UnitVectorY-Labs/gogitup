package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
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
	forceFlag := fs.Bool("force", false, "Refresh latest version from GitHub, ignoring cache")
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
	ghClient := github.NewDefaultClient(github.ResolveToken(cfg.GitHubAuth))

	entries := make([]checkEntry, 0, len(cfg.Apps))

	for _, app := range cfg.Apps {
		entry := checkEntry{Name: app.Name, InstalledVersion: "unknown", LatestVersion: "unknown"}

		info, err := runner.GetInfo(app.Name)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not get info for '%s': %v", app.Name, err))
			entries = append(entries, entry)
			continue
		}
		entry.InstalledVersion = info.Version

		owner, repo, err := goversion.ParseGitHubRepo(info.Path)
		if err != nil {
			output.Warn(fmt.Sprintf("Could not parse GitHub repo for '%s': %v", app.Name, err))
			entries = append(entries, entry)
			continue
		}

		// Check cache first unless --force is set.
		cached, found := cache.Get(c, app.Name)
		if !*forceFlag && found && !cache.IsExpired(cached, cache.DefaultTTL) {
			entry.LatestVersion = cached.LatestVersion
		} else {
			latest, err := ghClient.GetLatestRelease(owner, repo)
			if err != nil {
				output.Warn(fmt.Sprintf("Could not fetch latest release for '%s': %v", app.Name, err))
				entries = append(entries, entry)
				continue
			}
			entry.LatestVersion = latest
			cache.Set(c, app.Name, latest)
		}

		entry.UpdateAvailable = entry.InstalledVersion != entry.LatestVersion
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
}
