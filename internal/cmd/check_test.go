package cmd

import (
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func TestCheckGoVersionColorGreenWhenCurrent(t *testing.T) {
	e := checkEntry{GoVersion: "1.25.7", goVersionRaw: "go1.25.7", GoVersionNewer: false}
	if got := checkGoVersionColor(e); got != output.Green {
		t.Fatalf("expected green for current toolchain, got %q", got)
	}
}

func TestCheckGoVersionColorYellowWhenToolchainNewer(t *testing.T) {
	e := checkEntry{GoVersion: "1.25.7", goVersionRaw: "go1.25.7", GoVersionNewer: true}
	if got := checkGoVersionColor(e); got != output.Yellow {
		t.Fatalf("expected yellow when toolchain is newer, got %q", got)
	}
}

func TestCheckGoVersionColorGrayWhenUnknown(t *testing.T) {
	e := checkEntry{GoVersion: "unknown", goVersionRaw: ""}
	if got := checkGoVersionColor(e); got != output.Gray {
		t.Fatalf("expected gray for unknown go version, got %q", got)
	}
}
