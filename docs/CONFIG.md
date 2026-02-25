---
layout: default
title: Config
nav_order: 4
permalink: /config
---

# Config
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Config File

The configuration file is located at `~/.gogitup` and uses YAML format. It is created automatically the first time you register a binary.

### Example

```yaml
apps:
  - name: golangci-lint
  - name: goimports
github_auth: false
```

### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `apps` | list | `[]` | List of registered application binary names |
| `apps[].name` | string | - | Binary name of the registered application |
| `github_auth` | boolean | `false` | Enable authenticated GitHub API requests |

## GitHub Authentication

When `github_auth` is set to `true`, gogitup sends authenticated requests to the GitHub API. This is useful for avoiding rate limits.

The token is resolved in the following order:

1. The `GITHUB_TOKEN` environment variable, if set.
2. The output of `gh auth token` (GitHub CLI), as a fallback.

If neither source provides a token, requests are made without authentication.

## Cache File

The cache file is located at `~/.gogitup.cache` and uses YAML format. It stores the latest version information retrieved from GitHub so that repeated checks do not require additional API calls.

Cache entries expire after **24 hours**. After expiry the next `check` or `update` will re-fetch the latest release from GitHub.

### Example

```yaml
entries:
  golangci-lint:
    latest_version: v1.62.2
    checked_at: 2025-01-15T10:30:00Z
```
