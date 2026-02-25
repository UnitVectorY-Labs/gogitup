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
  - name: ghorgsync
  - name: bulkfilepr
github_auth: false
```

### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `apps` | list | `[]` | List of registered application binary names |
| `apps[].name` | string | - | Binary name of the registered application |
| `github_auth` | boolean | `false` | Enable authenticated GitHub API requests |

## GitHub Authentication

When `github_auth` is set to `true`, **gogitup** sends authenticated requests to the GitHub API. This is useful for avoiding rate limits. By default, **gogitup** does not authenticate and is subject to GitHub's unauthenticated rate limits.

When enabled, the token is resolved in the following order:

1. The `GITHUB_TOKEN` environment variable, if set.
2. The output of `gh auth token` (GitHub CLI), as a fallback.

If neither source provides a token, requests are made without authentication.

## Cache File

The cache file is located at `~/.gogitup.cache` and uses YAML format. It stores the latest version information retrieved from GitHub so that repeated checks do not require additional API calls.

{: .important }
Cache entries expire after **24 hours**. After expiry the next `check` or `update` will re-fetch the latest release from GitHub. A check can be forced with `--force` to bypass the cache, but the purpose of the cache is to avoid unnecessary API calls to GitHub.

### Example

```yaml
entries:
  ghorgsync:
    latest_version: v0.10.0
    checked_at: 2025-01-15T10:30:00Z
```
