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

## Listing Tracked Tools

View all registered binaries and their installed versions:

```bash
gogitup list
```

```
Registered Binaries

  Name        Module Path                             Installed Version
  ──────────  ──────────────────────────────────────  ─────────────────
  bulkfilepr  github.com/UnitVectorY-Labs/bulkfilepr  v0.2.2           
  ghorgsync   github.com/UnitVectorY-Labs/ghorgsync   v0.1.0           
```

## Checking for Updates

See which tools have newer releases available on GitHub:

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
The `check` command only checks for updates once every 24 hours caching the latest release information. Use `gogitup check --force` to bypass the cache and re-fetch from GitHub.

## Updating All Tools

Run a single command to update every tracked binary that has a newer release:

```bash
gogitup update
```

```
gogitup update
⟳ Updating 'bulkfilepr' from v0.2.2 to v0.2.3...
✓ Updated 'bulkfilepr' to v0.2.3
ℹ 'ghorgsync' is already up to date (v0.1.0)

✓ Updated 1 binary(ies).
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

## Enabling GitHub Authentication

Edit `~/.gogitup` and set `github_auth` to `true` to use authenticated API requests and avoid rate limits:

```yaml
apps:
  - name: ghorgsync
  - name: bulkfilepr
github_auth: true
```

Then ensure a token is available via the `GITHUB_TOKEN` environment variable or the GitHub CLI (`gh auth token`).
