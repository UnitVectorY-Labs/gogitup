package cmd

import (
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func runAdd(args []string) {
	if len(args) < 1 {
		output.Error("Usage: gogitup add <binary-name>")
		os.Exit(1)
	}

	name := args[0]

	runner := &goversion.DefaultRunner{}
	info, err := runner.GetInfo(name)
	if err != nil {
		output.Error(fmt.Sprintf("Cannot find binary '%s': %v", name, err))
		os.Exit(1)
	}

	if !goversion.IsGitHubRepo(info.Path) {
		output.Error(fmt.Sprintf("Module path '%s' is not a GitHub repository", info.Path))
		os.Exit(1)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	if err := config.AddApp(cfg, name); err != nil {
		output.Warn(err.Error())
		os.Exit(1)
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.Success(fmt.Sprintf("Added '%s' (%s)", name, info.Path))
}
