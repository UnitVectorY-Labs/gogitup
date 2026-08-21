package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type listEntry struct {
	Name             string `json:"name"`
	ModulePath       string `json:"module_path"`
	InstalledVersion string `json:"installed_version"`
	GoVersion        string `json:"go_version"`
	goVersionRaw     string
}

func collectListEntries(apps []config.App, runner goversion.Runner) []listEntry {
	entries := make([]listEntry, 0, len(apps))
	for _, app := range apps {
		entry := listEntry{Name: app.Name, ModulePath: "unknown", InstalledVersion: "unknown", GoVersion: "unknown"}
		info, err := runner.GetInfo(app.Name)
		if err == nil {
			entry.ModulePath = info.Path
			entry.InstalledVersion = info.Version
			entry.goVersionRaw = info.GoVersion
			entry.GoVersion = strings.TrimPrefix(info.GoVersion, "go")
			if info.GoVersion == "" {
				entry.GoVersion = "unknown"
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	jsonFlag := fs.Bool("json", false, "Output as JSON")
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

	runner := &goversion.DefaultRunner{}
	entries := collectListEntries(cfg.Apps, runner)

	if *jsonFlag {
		if err := output.PrintJSON(entries); err != nil {
			output.Error(fmt.Sprintf("Failed to output JSON: %v", err))
			os.Exit(1)
		}
		return
	}

	activeGoVersion, err := goversion.CurrentToolchainVersion()
	if err != nil {
		activeGoVersion = ""
	}

	output.Header("Registered Binaries")
	printListTable(os.Stdout, entries, activeGoVersion)
}

func printListTable(w io.Writer, entries []listEntry, activeGoVersion string) {
	// Calculate column widths. The two version headings are stacked to keep the
	// table compact while preserving alignment with their data columns.
	nameW, pathW, verW, goW := len("Name"), len("Module Path"), len("Installed"), len("Version")
	for _, e := range entries {
		if len(e.Name) > nameW {
			nameW = len(e.Name)
		}
		if len(e.ModulePath) > pathW {
			pathW = len(e.ModulePath)
		}
		if len(e.InstalledVersion) > verW {
			verW = len(e.InstalledVersion)
		}
		if len(e.GoVersion) > goW {
			goW = len(e.GoVersion)
		}
	}

	fmt.Fprintln(w)
	// Header rows
	fmt.Fprintf(w, "  %s%s%-*s  %-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "", pathW, "", verW, "Installed", goW, "Go", output.Reset)
	fmt.Fprintf(w, "  %s%s%-*s  %-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "Name", pathW, "Module Path", verW, "Version", goW, "Version", output.Reset)
	// Separator
	fmt.Fprintf(w, "  %s%s  %s  %s  %s%s\n", output.Gray,
		strings.Repeat("─", nameW), strings.Repeat("─", pathW), strings.Repeat("─", verW), strings.Repeat("─", goW), output.Reset)
	// Data rows
	for _, e := range entries {
		fmt.Fprintf(w, "  %-*s  %s%-*s%s  %s%-*s%s  %s%-*s%s\n",
			nameW, e.Name,
			output.Gray, pathW, e.ModulePath, output.Reset,
			output.Green, verW, e.InstalledVersion, output.Reset,
			listGoVersionColor(e, activeGoVersion), goW, e.GoVersion, output.Reset)
	}
	fmt.Fprintln(w)
}

func listGoVersionColor(entry listEntry, activeGoVersion string) string {
	if entry.goVersionRaw == "" {
		return output.Gray
	}
	if activeGoVersion == "" {
		return ""
	}
	if entry.goVersionRaw == activeGoVersion {
		return output.Green
	}
	return output.Red
}
