---
layout: default
title: Examples
nav_order: 5
permalink: /examples
---

# Examples
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Adding a Tool

Register a Go-installed binary so gogitup can track it:

```bash
gogitup add ghorgsync
```

Install and register a tool directly from a full Go command package path:

```bash
gogitup install golang.org/x/vuln/cmd/govulncheck
```

Install and register a tool from a private GitHub repository:

```bash
gh auth login
gogitup install --private owner/private-tool
```

`GITHUB_TOKEN` may be used instead of GitHub CLI authentication. The private marker is saved with the application, so `gogitup check` and `gogitup upgrade` continue to use authenticated private-repository access automatically.

## Listing Tracked Tools

View all registered binaries, their installed versions, and the Go versions used to build them:

```bash
gogitup list
```

```
Registered Binaries

                                                      Installed  Go
  Name        Module Path                             Version    Version
  ──────────  ──────────────────────────────────────  ─────────  ───────
  bulkfilepr  github.com/UnitVectorY-Labs/bulkfilepr  v0.2.2     1.25.7
  ghorgsync   github.com/UnitVectorY-Labs/ghorgsync   v0.1.0     1.25.7
```

Go build versions matching the active locally installed toolchain are shown in green; versions that differ are shown in red.

## Checking for Updates

See which tools have newer versions available:

```bash
gogitup check
```

```
Update Check

  Name        Installed  Latest  Update
  ──────────  ─────────  ──────  ──────
  bulkfilepr  v0.2.2     v0.2.3  yes   
  ghorgsync   v0.1.0     v0.1.0  no    
```

{: .important }
The `check` command only checks for updates once every 24 hours, caching the latest version information. Use `gogitup check --force` to bypass the cache and fetch it again.

## Upgrading All Tools

Run a single command to upgrade every tracked binary that has a newer release:

```bash
gogitup upgrade
```

```
gogitup upgrade --verbose
⟳ Upgrading 'bulkfilepr' from v0.2.2 to v0.2.3...
✓ Upgraded 'bulkfilepr' to v0.2.3
ℹ 'ghorgsync' is already up to date (v0.1.0)

✓ Upgraded 1 binary(ies).
```

To also rebuild otherwise-current binaries that were compiled with an older Go version:

```bash
gogitup upgrade --go-version
```

For example, if `tool` v1.2.3 was built with Go 1.25.6 and the active `go` command is Go 1.25.7, **gogitup** reinstalls `tool@v1.2.3` directly. The existing binary is overwritten; it does not need to be uninstalled first.

Preview upgrades without installing anything or changing the cache:

```bash
gogitup upgrade --dry-run
```

Include Go-version rebuilds in the preview:

```bash
gogitup upgrade --dry-run --go-version
```

## Using JSON Output for Scripting

Both `list` and `check` support `--json` for machine-readable output:

```bash
gogitup check --json
```

```json
[
  {
    "name": "bulkfilepr",
    "installed_version": "v0.2.2",
    "latest_version": "v0.2.3",
    "update_available": true
  },
  {
    "name": "ghorgsync",
    "installed_version": "v0.1.0",
    "latest_version": "v0.1.0",
    "update_available": false
  }
]
```

## Removing a Tool

Stop tracking a binary without uninstalling it:

```bash
gogitup remove ghorgsync
```

Stop tracking a binary and delete its executable from `PATH`:

```bash
gogitup remove --delete ghorgsync
```

## Enabling GitHub Authentication

Edit `~/.gogitup` and set `github_auth` to `true` to use authenticated API requests and avoid rate limits:

```yaml
apps:
  - name: ghorgsync
  - name: bulkfilepr
github_auth: true
```

Then ensure a token is available via the `GITHUB_TOKEN` environment variable or the GitHub CLI (`gh auth token`).
