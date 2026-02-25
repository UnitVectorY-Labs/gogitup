package cmd

import (
	"flag"
	"fmt"
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
	entries := make([]listEntry, 0, len(cfg.Apps))

	for _, app := range cfg.Apps {
		entry := listEntry{Name: app.Name, ModulePath: "unknown", InstalledVersion: "unknown"}
		info, err := runner.GetInfo(app.Name)
		if err == nil {
			entry.ModulePath = info.Path
			entry.InstalledVersion = info.Version
		}
		entries = append(entries, entry)
	}

	if *jsonFlag {
		if err := output.PrintJSON(entries); err != nil {
			output.Error(fmt.Sprintf("Failed to output JSON: %v", err))
			os.Exit(1)
		}
		return
	}

	// Calculate column widths
	nameW, pathW, verW := len("Name"), len("Module Path"), len("Installed Version")
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
	}

	output.Header("Registered Binaries")
	fmt.Println()
	// Header row
	fmt.Printf("  %s%s%-*s  %-*s  %-*s%s\n", output.Bold, output.Cyan,
		nameW, "Name", pathW, "Module Path", verW, "Installed Version", output.Reset)
	// Separator
	fmt.Printf("  %s%s  %s  %s%s\n", output.Gray,
		strings.Repeat("─", nameW), strings.Repeat("─", pathW), strings.Repeat("─", verW), output.Reset)
	// Data rows
	for _, e := range entries {
		fmt.Printf("  %-*s  %s%-*s%s  %s%-*s%s\n",
			nameW, e.Name,
			output.Gray, pathW, e.ModulePath, output.Reset,
			output.Green, verW, e.InstalledVersion, output.Reset)
	}
	fmt.Println()
}

