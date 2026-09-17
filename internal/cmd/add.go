package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type addOptions struct {
	Name       string
	GitHubUser string
}

func parseAddOptions(args []string) (addOptions, error) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	user := fs.String("github-user", "", "GitHub CLI account to use for this app")
	if err := fs.Parse(args); err != nil {
		return addOptions{}, err
	}
	if fs.NArg() != 1 {
		return addOptions{}, fmt.Errorf("usage: gogitup add [--github-user USER] <binary-name>")
	}
	if err := config.ValidateGitHubUser(*user); err != nil {
		return addOptions{}, err
	}
	return addOptions{Name: fs.Arg(0), GitHubUser: *user}, nil
}

func runAdd(args []string) {
	opts, err := parseAddOptions(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(output.DefaultWriter.Out, "Usage: gogitup add [--github-user USER] <binary-name>")
			fmt.Fprintln(output.DefaultWriter.Out, "  --github-user USER  GitHub CLI account to use for this app")
			return
		}
		output.Error(err.Error())
		os.Exit(2)
	}
	name := opts.Name

	runner := &goversion.DefaultRunner{}
	info, err := runner.GetInfo(name)
	if err != nil {
		output.Error(fmt.Sprintf("Cannot find binary '%s': %v", name, err))
		os.Exit(1)
	}

	if err := validateAppGitHub(config.App{GitHubUser: opts.GitHubUser}, info.Path); err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	installPath := ""
	if !goversion.IsGitHubRepo(info.Path) {
		installPath = info.PackagePath
	}
	if err := config.RegisterApp(cfg, config.App{Name: name, InstallPath: installPath, GitHubUser: opts.GitHubUser}); err != nil {
		output.Warn(err.Error())
		os.Exit(1)
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.Success(fmt.Sprintf("Added '%s' (%s)", name, info.Path))
}
