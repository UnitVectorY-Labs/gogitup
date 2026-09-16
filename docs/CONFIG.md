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
  - name: private-tool
    private: true
    github_user: jared-work
github_auth: false
goproxy: "https://proxy.golang.org"
```

### Attributes

| Attribute | Type | Default | Description |
|-----------|------|---------|-------------|
| `apps` | list | `[]` | List of registered application binary names |
| `apps[].name` | string | - | Binary name of the registered application |
| `apps[].install_path` | string | `""` | Optional Go command package path used for upgrades |
| `apps[].private` | boolean | `false` | Treat the application's GitHub repository as private during checks and upgrades |
| `apps[].github_user` | string | `""` | Stored GitHub CLI account to use for this app, independently of repository owner |
| `github_auth` | boolean | `false` | Enable authenticated GitHub API requests for apps without an explicit account |
| `goproxy` | string | `""` | Override the `GOPROXY` environment variable used when running `go install` |
| `cgo_enabled` | boolean | (inherited) | Override the `CGO_ENABLED` environment variable used when running `go install` |

## GitHub Authentication

When `github_auth` is set to `true`, **gogitup** sends authenticated requests to the GitHub API. This is useful for avoiding rate limits. By default, public repositories use unauthenticated requests and are subject to GitHub's unauthenticated rate limits.

Applications installed with `gogitup install --private` are stored with `private: true`. Private applications always use authenticated GitHub requests during `check` and `upgrade`, regardless of the global `github_auth` value. Their `go install` processes also receive repository-scoped private-module and Git authentication settings. These settings apply only to the child process; **gogitup** does not modify persistent Go or Git configuration.

For apps without `github_user`, authentication is resolved independently for each app when `github_auth` or that app's `private` setting is enabled. The token is resolved in the following order:

1. The `GITHUB_TOKEN` environment variable, if set.
2. The output of `gh auth token` (GitHub CLI), as a fallback.

If neither source provides a token, public repository requests are made without authentication. Private installs, checks, and upgrades instead fail with an authentication error. A private app does not enable authentication for other public apps.

### Per-App GitHub Accounts

Set `apps[].github_user` to select an account already logged in through GitHub CLI. The account can differ from the repository owner and from the account used by other apps:

```yaml
apps:
  - name: internal-tool
    private: true
    github_user: jared-work
  - name: public-tool
    github_user: jared-personal
  - name: govulncheck
    install_path: golang.org/x/vuln/cmd/govulncheck
github_auth: false
```

`gogitup install --github-user USER <path>` and `gogitup add --github-user USER <name>` save this setting on a new registration. To change an existing registration, edit its `github_user` in `~/.gogitup`; remove the field or set it to `""` to restore default authentication. Duplicate registrations retain their existing settings. Usernames must not contain whitespace or control characters.

For an app with an explicit account, **gogitup** reads its token with `gh auth token --hostname github.com --user USER`. This authenticates release requests even with `github_auth: false`. **gogitup** never switches the shared active `gh` account, so there is no switch-back option or restoration step. Tokens are not saved in configuration or cache files.

Account selection and repository privacy are separate. Set `private: true` (or install with `--private`) when private source downloads are needed. The selected account's token then also supplies the existing repository-scoped HTTPS Git authentication in the child `go install` process. SSH keys, URL rewrites, and credentials for other repositories or private dependencies remain governed by your Git and Go configuration.

Both `github_user` and private repository support are limited to `github.com` app modules. Other module hosts and vanity module paths are not supported by these options.

When an explicit account needs credentials, nonempty `GH_TOKEN` and `GITHUB_TOKEN` are rejected with instructions to unset them. There is no fallback to environment tokens, another account, or anonymous access. Missing GitHub CLI support, a missing stored token, and failed API access are reported with the app and account names. Log in manually using `gh auth login --hostname github.com`; **gogitup** does not perform interactive login or check credentials during `add` or `list`.

`check` and `upgrade` continue with other apps after an app fails and return exit status `1` for failed app operations. Errors go to stderr, leaving `check --json` output as valid JSON. Failed checks report an unknown latest version and are not cached. For private apps or apps with an explicit account, any failed release lookup prevents a fallback `--go-version` rebuild, including during `--dry-run`.

See the GitHub CLI manuals for [account token selection](https://cli.github.com/manual/gh_auth_token) and [environment token precedence](https://cli.github.com/manual/gh_help_environment).

## GOPROXY

When `goproxy` is set, **gogitup** passes the configured value as the `GOPROXY` environment variable when checking non-GitHub module updates with `go list -m -u` and when running `go install` (during both `install` and `upgrade`). This is useful in environments that require a custom module proxy.

{: .note }
If `goproxy` is not set or is empty, the `GOPROXY` value is inherited from the current process environment (the default Go behavior).

## CGO_ENABLED

When `cgo_enabled` is set, **gogitup** passes the configured value as the `CGO_ENABLED` environment variable when running `go install` (during both `install` and `upgrade`). Setting `cgo_enabled: false` disables cgo for all installs and updates, which is useful in environments where cgo is unavailable or undesirable.

{: .note }
If `cgo_enabled` is not set, the `CGO_ENABLED` value is inherited from the current process environment (the default Go behavior).

## Cache File

The cache file is located at `~/.gogitup.cache` and uses YAML format. It stores version-check results so repeated checks do not require additional GitHub or Go module proxy requests. Each result is associated with the installed version, configured GitHub account, and privacy setting that were checked. Changing or removing `github_user`, or changing `private`, causes a fresh lookup. Existing cache entries without these fields are treated as having no explicit account and `private: false`. Account names are compared case-insensitively for cache matching.

{: .important }
Cache entries expire after **24 hours**. After expiry, the next `check` refreshes the result from GitHub or the configured Go module proxy. The `upgrade` command always performs a fresh lookup. A check can be forced with `--force` to bypass the cache. A valid cache hit needs no credential lookup or GitHub CLI, even for a private app; cached results do not verify current access permissions.

### Example

```yaml
entries:
  ghorgsync:
    latest_version: v0.10.0
    installed_version: v0.9.0
    checked_at: 2025-01-15T10:30:00Z
```
