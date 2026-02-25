package cmd

import (
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

// Execute is the main entry point for the CLI. It parses os.Args to determine
// the subcommand and dispatches accordingly.
func Execute(version string) {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	subcmd := os.Args[1]

	switch subcmd {
	case "--version", "-v":
		fmt.Println("gogitup " + version)
	case "add":
		runAdd(os.Args[2:])
	case "remove":
		runRemove(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "check":
		runCheck(os.Args[2:])
	case "update":
		runUpdate(os.Args[2:])
	case "--help", "-h", "help":
		printHelp()
	default:
		output.Error("Unknown command: " + subcmd)
		fmt.Println()
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	output.Header("gogitup - Keep your Go-installed binaries up to date")
	fmt.Println()
	fmt.Printf("  %sUsage:%s gogitup <command> [arguments]\n", output.Bold, output.Reset)
	fmt.Println()
	fmt.Printf("  %sCommands:%s\n", output.Bold, output.Reset)
	fmt.Printf("    %sadd%s <name>       Register a Go-installed binary\n", output.Cyan, output.Reset)
	fmt.Printf("    %sremove%s <name>    Remove a registered binary\n", output.Cyan, output.Reset)
	fmt.Printf("    %slist%s             List registered binaries and installed versions\n", output.Cyan, output.Reset)
	fmt.Printf("    %scheck%s            Check for available updates\n", output.Cyan, output.Reset)
	fmt.Printf("    %supdate%s           Update all binaries with available updates\n", output.Cyan, output.Reset)
	fmt.Println()
	fmt.Printf("  %sFlags:%s\n", output.Bold, output.Reset)
	fmt.Printf("    %s--version, -v%s    Print version\n", output.Cyan, output.Reset)
	fmt.Printf("    %s--help, -h%s       Show this help message\n", output.Cyan, output.Reset)
	fmt.Println()
}
