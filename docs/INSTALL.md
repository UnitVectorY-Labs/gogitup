---
layout: default
title: Installation
nav_order: 2
permalink: /install
---

# Installation
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Installation Methods

There are several ways to install **gogitup**:

### Download Binary

Download pre-built binaries from the [GitHub Releases](https://github.com/UnitVectorY-Labs/gogitup/releases) page for the latest version.

[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/gogitup.svg)](https://github.com/UnitVectorY-Labs/gogitup/releases/latest) 

Choose the appropriate binary for your platform and add it to your PATH.

### Install Using Go

Install directly from the Go toolchain:

```bash
go install github.com/UnitVectorY-Labs/gogitup@latest
```

### Build from Source

Build the application from source code:

```bash
git clone https://github.com/UnitVectorY-Labs/gogitup.git
cd gogitup
go build -o gogitup
```

## Optional GitHub Account Selection

Per-app `github_user` requires a GitHub CLI version supporting `gh auth token --hostname github.com --user USER`. Log in to each account with `gh auth login --hostname github.com` before using it. Unset `GH_TOKEN` and `GITHUB_TOKEN` for operations using an explicit account. See [per-app GitHub accounts](CONFIG.md#per-app-github-accounts) for configuration and examples.

## Upgrading with gogitup

### Registering with gogitup

Tell **gogitup** that it should manage the updates for **gogitup**:

```bash
gogitup add gogitup
```

### Upgrading with gogitup

Upgrade the registered packages using **gogitup**:

```bash
gogitup upgrade
```
