package cmd

import (
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func runRemove(args []string) {
	if len(args) < 1 {
		output.Error("Usage: gogitup remove <binary-name>")
		os.Exit(1)
	}

	name := args[0]

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	if err := config.RemoveApp(cfg, name); err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	cachePath := cache.DefaultPath()
	c, err := cache.Load(cachePath)
	if err == nil {
		cache.Remove(c, name)
		_ = cache.Save(cachePath, c)
	}

	output.Success(fmt.Sprintf("Removed '%s'", name))
}
