# gogitup Agent Notes

This is a Go based command line application the follows idiomatic Go conventions. Its purpose is to sync a local directory of git repositories with the corresponding repositories on GitHub. The main.go file in the root provides the entry point, all other code is organized into packages under the `internal/` directory. The `docs/` directory contains markdown files that document the behavior and configuration of the application, and should be updated in tandem with code changes to ensure they remain accurate.

## Repo Conventions For The Agent
- Treat `docs/` as required source of truth alongside code. When behavior changes, update the matching doc pages in the same PR.
- Keep dependencies minimal and prefer stdlib. Add third-party packages only when they materially improve correctness or output quality.

## Testing and Validation

- Never attempt to run go install or interact with GitHub API in tests. Instead, mock these interactions by simulating expected outputs and errors in unit tests of the relevant functions.
