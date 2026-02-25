---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## add

Registers a binary for tracking. The binary must already be installed via `go install` and must originate from a `github.com` module path.

```bash
gogitup add <name>
```

```bash
gogitup add golangci-lint
```

## remove

Removes a binary from tracking. The binary itself is not uninstalled.

```bash
gogitup remove <name>
```

```bash
gogitup remove golangci-lint
```

## list

Lists all registered binaries along with their currently installed versions.

```bash
gogitup list
```

Use `--json` to output the list as JSON, useful for scripting.

```bash
gogitup list --json
```

## check

Checks GitHub for newer releases of all registered binaries. Displays the installed version alongside the latest available version. Results are cached for 24 hours.

```bash
gogitup check
```

Use `--json` to output the results as JSON.

```bash
gogitup check --json
```

## update

Checks for updates and runs `go install` to update every registered binary that has a newer release available.

```bash
gogitup update
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--version`, `-v` | Print the gogitup version |
| `--help`, `-h` | Show help message |
