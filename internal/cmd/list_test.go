package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/UnitVectorY-Labs/gogitup/internal/config"
	"github.com/UnitVectorY-Labs/gogitup/internal/goversion"
	"github.com/UnitVectorY-Labs/gogitup/internal/output"
)

func TestCollectListEntriesIncludesGoVersion(t *testing.T) {
	runner := &stubRunner{
		infos: map[string]*goversion.Info{
			"tool": {Path: "github.com/acme/tool", Version: "v1.2.3", GoVersion: "go1.25.7"},
		},
		errs: map[string]error{"missing": errors.New("not found")},
	}

	entries := collectListEntries([]config.App{{Name: "tool"}, {Name: "missing"}}, runner)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].GoVersion != "1.25.7" {
		t.Fatalf("expected embedded Go version, got %#v", entries[0])
	}
	if entries[0].goVersionRaw != "go1.25.7" {
		t.Fatalf("expected raw embedded Go version, got %#v", entries[0])
	}
	if entries[1].GoVersion != "unknown" {
		t.Fatalf("expected unknown Go version, got %#v", entries[1])
	}
}

func TestListGoVersionColorComparesActiveLocalToolchain(t *testing.T) {
	entry := listEntry{GoVersion: "1.25.7", goVersionRaw: "go1.25.7"}

	if got := listGoVersionColor(entry, "go1.25.7"); got != output.Green {
		t.Fatalf("expected matching version to be green, got %q", got)
	}
	if got := listGoVersionColor(entry, "go1.25.8"); got != output.Red {
		t.Fatalf("expected mismatched version to be red, got %q", got)
	}
	if got := listGoVersionColor(entry, ""); got != "" {
		t.Fatalf("expected no status color without an active version, got %q", got)
	}
}

func TestPrintListTableUsesAlignedTwoLineVersionHeadings(t *testing.T) {
	entries := []listEntry{{
		Name:             "tool",
		ModulePath:       "example.com/tool",
		InstalledVersion: "v1.2.3",
		GoVersion:        "1.25.7",
		goVersionRaw:     "go1.25.7",
	}}
	var buf bytes.Buffer

	printListTable(&buf, entries, "go1.25.7")
	plain := buf.String()
	for _, code := range []string{output.Bold, output.Cyan, output.Gray, output.Green, output.Red, output.Reset} {
		plain = strings.ReplaceAll(plain, code, "")
	}
	lines := strings.Split(plain, "\n")
	if len(lines) < 5 {
		t.Fatalf("expected table output, got %q", plain)
	}
	if lines[1] != "                          Installed  Go     " {
		t.Fatalf("unexpected first header row %q", lines[1])
	}
	if lines[2] != "  Name  Module Path       Version    Version" {
		t.Fatalf("unexpected second header row %q", lines[2])
	}
	if !strings.HasPrefix(lines[4], "  tool  example.com/tool  v1.2.3") {
		t.Fatalf("data row is not aligned with headings: %q", lines[4])
	}
}

func TestCollectListEntriesTreatsEmptyGoVersionAsUnknown(t *testing.T) {
	runner := &stubRunner{infos: map[string]*goversion.Info{
		"tool": {Path: "github.com/acme/tool", Version: "v1.2.3"},
	}}

	entries := collectListEntries([]config.App{{Name: "tool"}}, runner)
	if entries[0].GoVersion != "unknown" {
		t.Fatalf("expected unknown Go version, got %#v", entries[0])
	}
}
