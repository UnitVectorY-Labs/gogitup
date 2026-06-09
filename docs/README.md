---
layout: default
title: gogitup
nav_order: 1
permalink: /
---

# gogitup

You installed a handful of Go tools with `go install`, but now they're out of date. Checking each project for a new version, then re-running `go install` for every one of them, gets old fast.

**gogitup** automates that. It tracks the Go-installed binaries you care about, checks for newer versions, and runs `go install` to bring them up to date with a single easy to use command.

## Quick Start

```bash
# Install gogitup
go install github.com/UnitVectorY-Labs/gogitup@latest

# Install and register a Go tool
gogitup install golang.org/x/vuln/cmd/govulncheck

# Check for updates
gogitup check

# Upgrade all tracked tools
gogitup upgrade
```
