package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/github"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/installer"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

type installDependencies struct {
	ghClient    github.Client
	installer   installer.Installer
	runner      goversion.Runner
	private     bool
	githubToken string
	out         *output.Writer
	errOut      *output.Writer
}

type installOptions struct {
	Private bool
	Target  string
}

func parseInstallOptions(args []string, stderr io.Writer) (installOptions, error) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	privateFlag := fs.Bool("private", false, "Install from a private GitHub repository")
	if err := fs.Parse(args); err != nil {
		return installOptions{}, err
	}
	if fs.NArg() != 1 {
		return installOptions{}, fmt.Errorf("usage: gogitup install [--private] <owner/repo|package-path>")
	}
	return installOptions{Private: *privateFlag, Target: fs.Arg(0)}, nil
}

func runInstall(args []string) {
	opts, err := parseInstallOptions(args, output.ErrorWriter.Out)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		output.Error(err.Error())
		os.Exit(2)
	}

	target, err := parseInstallTarget(opts.Target)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}
	if opts.Private && target.owner == "" {
		output.Error("--private is only supported for github.com repositories")
		os.Exit(1)
	}

	cfgPath := config.DefaultPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		output.Error(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	githubToken := github.ResolveToken(cfg.GitHubAuth || opts.Private)
	if opts.Private && githubToken == "" {
		output.Error("Private GitHub installation requires authentication; set GITHUB_TOKEN or run 'gh auth login'")
		os.Exit(1)
	}
	ghClient := github.NewDefaultClient(githubToken)
	inst := installer.NewDefaultInstallerWithOptions(cfg.GOPROXY, cfg.CGOEnabled)
	runner := &goversion.DefaultRunner{}

	deps := installDependencies{
		ghClient:    ghClient,
		installer:   inst,
		runner:      runner,
		private:     opts.Private,
		githubToken: githubToken,
		out:         output.DefaultWriter,
		errOut:      output.ErrorWriter,
	}

	binaryName, err := runInstallTarget(target, deps)
	if err != nil {
		output.Error(err.Error())
		os.Exit(1)
	}

	if err := config.AddAppWithInstallOptions(cfg, binaryName, target.installPath(), opts.Private); err != nil {
		output.Warn(err.Error())
		return
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		output.Error(fmt.Sprintf("Failed to save config: %v", err))
		os.Exit(1)
	}

	output.Success(fmt.Sprintf("Added '%s' to tracking", binaryName))
}

type installTarget struct {
	packagePath string
	owner       string
	repo        string
}

func (t installTarget) installPath() string {
	if t.owner == "" {
		return t.packagePath
	}
	if t.packagePath == t.modulePath() {
		return ""
	}
	return t.packagePath
}

func (t installTarget) modulePath() string {
	if t.owner == "" {
		return t.packagePath
	}
	return "github.com/" + t.owner + "/" + t.repo
}

// parseInstallTarget accepts the existing owner/repo forms and full Go package paths.
func parseInstallTarget(value string) (installTarget, error) {
	if packagePath, version, found := strings.Cut(value, "@"); found {
		if version != "latest" || strings.Contains(packagePath, "@") {
			return installTarget{}, fmt.Errorf("invalid install target: %q (only @latest is supported)", value)
		}
		value = packagePath
	}
	if value == "" || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") {
		return installTarget{}, fmt.Errorf("invalid install target: %q", value)
	}

	if strings.HasPrefix(value, "github.com/") {
		parts := strings.Split(value, "/")
		if len(parts) >= 3 && parts[1] != "" && parts[2] != "" {
			return installTarget{packagePath: value, owner: parts[1], repo: parts[2]}, nil
		}
		return installTarget{}, fmt.Errorf("invalid install target: %q", value)
	}

	parts := strings.Split(value, "/")
	if len(parts) == 2 && !strings.Contains(parts[0], ".") {
		owner, repo, err := parseOwnerRepo(value)
		if err != nil {
			return installTarget{}, err
		}
		return installTarget{
			packagePath: "github.com/" + owner + "/" + repo,
			owner:       owner,
			repo:        repo,
		}, nil
	}
	if len(parts) < 2 || !strings.Contains(parts[0], ".") {
		return installTarget{}, fmt.Errorf("invalid install target: %q (expected owner/repo or a full Go package path)", value)
	}

	return installTarget{packagePath: value}, nil
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
	return runInstallTarget(installTarget{
		packagePath: "github.com/" + owner + "/" + repo,
		owner:       owner,
		repo:        repo,
	}, deps)
}

// runInstallTarget installs from a GitHub release or resolves a non-GitHub package at latest.
func runInstallTarget(target installTarget, deps installDependencies) (string, error) {
	version := "latest"
	if target.owner != "" {
		var err error
		version, err = deps.ghClient.GetLatestRelease(target.owner, target.repo)
		if err != nil {
			return "", fmt.Errorf("failed to fetch latest release for %s/%s: %w", target.owner, target.repo, err)
		}
	}

	deps.out.StartProgress(fmt.Sprintf("Installing %s@%s", target.packagePath, version))

	installerOptions := []installer.InstallOptions(nil)
	if deps.private {
		installerOptions = append(installerOptions, installer.InstallOptions{
			PrivateModule: target.modulePath(),
			GitHubToken:   deps.githubToken,
		})
	}
	_, err := deps.installer.Install(target.packagePath, version, installerOptions...)
	if err != nil {
		return "", fmt.Errorf("installation failed: %w", err)
	}

	deps.out.Success(fmt.Sprintf("Installed %s@%s", target.packagePath, version))

	parts := strings.Split(target.packagePath, "/")
	binaryName := parts[len(parts)-1]
	if _, err := deps.runner.GetInfo(binaryName); err != nil {
		return "", fmt.Errorf("binary %q not found on PATH after install; use 'gogitup add <name>' to track it manually", binaryName)
	}

	return binaryName, nil
}
