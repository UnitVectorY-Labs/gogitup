package main

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/UnitVectorY-Labs/gogitup/internal/cmd"
)

var Version = "dev"

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

func buildVersionOutput(version string) string {
	normalized := version
	if semverRe.MatchString(normalized) && !strings.HasPrefix(normalized, "v") {
		normalized = "v" + normalized
	}
	return fmt.Sprintf("%s (%s, %s/%s)", normalized, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func main() {
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	cmd.Execute(buildVersionOutput(Version))
}
