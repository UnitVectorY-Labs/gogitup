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

**Global Flags**

| Flag | Description |
|------|-------------|
| `--version`, `-v` | Print the **gogitup** version |
| `--help`, `-h` | Show help message |

---

## `add`

Registers a binary for tracking with **gogitup**. The binary must already be installed via `go install`.

```bash
gogitup add <name>
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `<name>` | Yes | None | Binary name of the tool to track, must be available on `PATH` (for example `ghorgsync`) |

**What `add` does:**

`add` inspects the installed binary with `go version -m -json` to confirm it was installed with Go tooling and to extract its module and command package paths.

---

## `install`

Installs a Go binary and registers it with **gogitup** in a single step. Existing GitHub `owner/repo` inputs remain supported, and full Go command package paths can also be used.

```bash
gogitup install <owner/repo|package-path>
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `<owner/repo\|package-path>` | Yes | None | GitHub repository or full Go command package path |

```bash
gogitup install UnitVectorY-Labs/gogitup
gogitup install golang.org/x/vuln/cmd/govulncheck
```

**What `install` does:**

1. For GitHub repository and command package paths, fetches the latest GitHub release tag.
2. For a non-GitHub package path, uses `@latest`.
3. Verifies that the resulting binary (named after the final path component) is available on `PATH`.
4. Registers the binary with **gogitup** for future `check` and `upgrade` tracking.

An optional `@latest` suffix is accepted. Explicit version suffixes are not supported.

If the installed binary name differs from the repository name (uncommon), the installation itself still succeeds but the binary will not be registered automatically. Use `gogitup add <name>` to register it manually.

---

## `remove`

Removes a binary from tracking. The binary itself is not uninstalled, **gogitup** just stops tracking it for updates when you run `check` or `upgrade`.

```bash
gogitup remove <name>
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `<name>` | Yes | None | Registered binary name to remove |

---

## `list`

Lists all registered binaries along with their currently installed versions.

```bash
gogitup list [--json]
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `--json` | No | `false` | Output the list as JSON for scripting |

**What `list` does:**

`list` reads the tracked app names from `~/.gogitup` and inspects each installed binary with `go version -m -json` to report the installed version.

---

## `check`

Checks for newer versions of all registered binaries. GitHub modules use GitHub Releases. Other modules use `go list -m -u -json <module>@<installed-version>` so the Go toolchain determines whether a newer version is available.

```bash
gogitup check [--json] [--force]
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `--json` | No | `false` | Output the results as JSON |
| `--force` | No | `false` | Ignore cached latest-version values and fetch fresh version data |

**What `check` does:**

`gogitup check` determines update status by combining:

1. Installed binary metadata from `go version -m -json`.
2. The embedded module path.
3. GitHub Releases for GitHub modules, or the `Update` result from `go list -m -u -json <module>@<installed-version>` for other modules.
4. The local cache file `~/.gogitup.cache` (version-check results cached for 24 hours).

{: .important }
By default, `check` uses a non-expired cache entry to reduce remote lookups. Cached results are tied to the installed version that was checked; changing a binary outside **gogitup** causes a fresh lookup. Use `gogitup check --force` to bypass the cache and refresh the cached value immediately.

---

## `upgrade`

Checks for updates and runs `go install` to upgrade every registered binary that has a newer release available.

```bash
gogitup upgrade
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `--verbose` | No | `false` | Show binaries that are already up to date while checking for updates |

**What `upgrade` does:**

`upgrade` uses installed binary metadata (`go version -m -json`) and the appropriate version source to find an update, then runs `go install <package>@<version>` when one is available. For non-GitHub modules, the Go toolchain reports an update only when it considers a newer version available; a merely different version does not trigger an install or downgrade. For command packages below a module root, **gogitup** stores the original package path as an optional `install_path` value in `~/.gogitup`. When that value is absent, `upgrade` uses the command package path embedded in the binary, so existing name-only configuration entries remain valid.
