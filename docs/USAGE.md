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

Registers a binary for tracking with **gogitup**. The binary must already be installed via `go install` and must originate from a `github.com` module path.

```bash
gogitup add <name>
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `<name>` | Yes | None | Binary name of the tool to track, must be available on `PATH` (for example `ghorgsync`) |

**What `add` does:**

`add` inspects the installed binary with `go version -m -json` to confirm it was installed with Go tooling and to extract the embedded module path. The module path must resolve to a `github.com/<owner>/<repo>` repository.

---

## `remove`

Removes a binary from tracking. The binary itself is not uninstalled, **gogitup** just stops tracking it for updates when you run `check` or `update`.

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

Checks GitHub for newer releases of all registered binaries. Displays the installed version alongside the latest available version.

```bash
gogitup check [--json] [--force]
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| `--json` | No | `false` | Output the results as JSON |
| `--force` | No | `false` | Ignore cached latest-version values and fetch fresh release data from GitHub |

**What `check` does:**

`gogitup check` determines update status by combining:

1. Installed binary metadata from `go version -m -json`.
2. The embedded module path (must be a `github.com/<owner>/<repo>` module).
3. The GitHub Releases API (`/repos/<owner>/<repo>/releases/latest`) for the latest release tag.
4. The local cache file `~/.gogitup.cache` (latest release tags cached for 24 hours).

{: .important }
By default, `check` uses a non-expired cache entry to reduce GitHub API calls. Use `gogitup check --force` to bypass the cache and refresh the cached value immediately.

---

## `update`

Checks for updates and runs `go install` to update every registered binary that has a newer release available.

```bash
gogitup update
```

| Name | Required | Default | Description |
|------|----------|---------|-------------|
| _(none)_ | No | N/A | This command currently has no command-specific arguments |

**What `update` does:**

`update` uses installed binary metadata (`go version -m -json`) plus the GitHub Releases API to find the latest release, then runs `go install <module>@<tag>` when an update is available. It refreshes the cache with the latest fetched tag and always fetches fresh release data (it does not rely on cached latest-version values).
