package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/installer"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type installDependencies struct {
	ghClient  github.Client
	installer installer.Installer
	runner    goversion.Runner
	out       *output.Writer
	errOut    *output.Writer
}

func runInstall(args []string) {
	if len(args) < 1 {
		output.Error("Usage: gogitup install <owner/repo>")
		os.Exit(1)
	}

	ownerRepo := args[0]
	owner, repo, err := parseOwnerRepo(ownerRepo)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	ghClient := github.NewDefaultClient(github.ResolveToken(cfg.GitHubAuth))
	inst := installer.NewDefaultInstallerWithGOPROXY(cfg.GOPROXY)
	runner := &goversion.DefaultRunner{}

	deps := installDependencies{
		ghClient:  ghClient,
		installer: inst,
		runner:    runner,
		out:       output.DefaultWriter,
		errOut:    output.ErrorWriter,
	}

	binaryName, err := runInstallApp(owner, repo, deps)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if err := config.AddApp(cfg, binaryName); err != nil {
		output.Warn(err.Error())
		return
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.Success(fmt.Sprintf("Added '%s' to tracking", binaryName))
}

// parseOwnerRepo parses an owner/repo string (with or without the "github.com/" prefix)
// and returns the owner and repo components.
func parseOwnerRepo(ownerRepo string) (owner, repo string, err error) {
	s := ownerRepo
	if strings.HasPrefix(s, "github.com/") {
		s = s[len("github.com/"):]
	}
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %q (expected owner/repo)", ownerRepo)
	}
	return parts[0], parts[1], nil
}

// runInstallApp fetches the latest release for owner/repo, runs go install, and verifies
// the resulting binary is available on PATH. It returns the binary name on success.
func runInstallApp(owner, repo string, deps installDependencies) (string, error) {
	latest, err := deps.ghClient.GetLatestRelease(owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release for %s/%s: %w", owner, repo, err)
	}

	modulePath := "github.com/" + owner + "/" + repo

	deps.out.StartProgress(fmt.Sprintf("Installing %s@%s", modulePath, latest))

	_, err = deps.installer.Install(modulePath, latest)
	if err != nil {
		return "", fmt.Errorf("installation failed: %w", err)
	}

	deps.out.Success(fmt.Sprintf("Installed %s@%s", modulePath, latest))

	// The binary is conventionally named after the repository.
	binaryName := repo
	if _, err := deps.runner.GetInfo(binaryName); err != nil {
		return "", fmt.Errorf("binary %q not found on PATH after install; use 'gogitup add <name>' to track it manually", binaryName)
	}

	return binaryName, nil
}
