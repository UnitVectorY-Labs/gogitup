package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/UnitVectorY-Labs/gogitup/internal/cache"
	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func runRemove(args []string) {
	opts, err := parseRemoveArgs(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printRemoveHelp(output.DefaultWriter.Out)
			return
		}
		output.Error(err.Error())
		os.Exit(2)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	if !config.HasApp(cfg, opts.name) {
		output.Error("app not found: " + opts.name)
		os.Exit(1)
	}

	if opts.deleteBinary {
		if err := deleteInstalledBinary(opts.name); err != nil {
			output.Error(fmt.Sprintf("Failed to delete binary %q: %v", opts.name, err))
			os.Exit(1)
		}
	}

	if err := config.RemoveApp(cfg, opts.name); err != nil {
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
		cache.Remove(c, opts.name)
		_ = cache.Save(cachePath, c)
	}

	if opts.deleteBinary {
		output.Success(fmt.Sprintf("Removed '%s' and deleted its binary", opts.name))
		return
	}

	output.Success(fmt.Sprintf("Removed '%s'", opts.name))
}

type removeOptions struct {
	name         string
	deleteBinary bool
}

func parseRemoveArgs(args []string) (removeOptions, error) {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	deleteFlag := fs.Bool("delete", false, "Delete the installed binary from PATH")
	if err := fs.Parse(args); err != nil {
		return removeOptions{}, err
	}
	if fs.NArg() != 1 {
		return removeOptions{}, errors.New("usage: gogitup remove [--delete] <binary-name>")
	}
	return removeOptions{name: fs.Arg(0), deleteBinary: *deleteFlag}, nil
}

func printRemoveHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: gogitup remove [--delete] <binary-name>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --delete  Delete the installed binary from PATH")
}

// deleteInstalledBinary deletes the executable resolved from PATH for name.
func deleteInstalledBinary(name string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("binary not found on PATH: %s", name)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("resolved path is a directory: %s", path)
	}

	return os.Remove(path)
}
