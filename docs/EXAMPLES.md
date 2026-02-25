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
gogitup add golangci-lint
```

## Listing Tracked Tools

View all registered binaries and their installed versions:

```bash
gogitup list
```

```
golangci-lint  v1.61.0
goimports      v0.28.0
```

## Checking for Updates

See which tools have newer releases available on GitHub:

```bash
gogitup check
```

```
golangci-lint  v1.61.0 -> v1.62.2  (update available)
goimports      v0.28.0              (up to date)
```

## Updating All Tools

Run a single command to update every tracked binary that has a newer release:

```bash
gogitup update
```

## Using JSON Output for Scripting

Both `list` and `check` support `--json` for machine-readable output:

```bash
gogitup check --json
```

```json
[
  {
    "name": "golangci-lint",
    "installed_version": "v1.61.0",
    "latest_version": "v1.62.2",
    "update_available": true
  }
]
```

## Removing a Tool

Stop tracking a binary without uninstalling it:

```bash
gogitup remove goimports
```

## Enabling GitHub Authentication

Edit `~/.gogitup` and set `github_auth` to `true` to use authenticated API requests and avoid rate limits:

```yaml
apps:
  - name: golangci-lint
  - name: goimports
github_auth: true
```

Then ensure a token is available via the `GITHUB_TOKEN` environment variable or the GitHub CLI (`gh auth token`).

## Configuration with Multiple Tools

A typical `~/.gogitup` file tracking several tools:

```yaml
apps:
  - name: golangci-lint
  - name: goimports
  - name: dlv
  - name: gopls
github_auth: false
```