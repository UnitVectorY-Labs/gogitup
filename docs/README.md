---
layout: default
title: gogitup
nav_order: 1
permalink: /
---

# gogitup

You installed a handful of Go tools with `go install`, but now they're out of date. Checking each project's GitHub page for a new release, then re-running `go install` for every one of them, gets old fast.

**gogitup** automates that. It tracks the Go-installed binaries you care about, checks GitHub for newer releases, and runs `go install` to bring them up to date — all with a single command.

## Quick Start

```bash
# Install gogitup
go install github.com/UnitVectorY-Labs/gogitup@latest

# Register a tool
gogitup add golangci-lint

# Update all tracked tools
gogitup update
```

For detailed usage, configuration, and examples see the rest of the documentation.