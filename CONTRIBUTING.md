# Contributing to lazyjira

Thanks for your interest in contributing!

## Getting started

```bash
git clone https://github.com/textfuel/lazyjira.git
cd lazyjira
make build
```

## Development workflow

1. Fork the repo and create a feature branch
2. Make your changes
3. Run `make check` (lint + vet + build) - this must pass
4. Open a pull request against `main`

## Git hooks

Install [mise](https://mise.jdx.dev/getting-started.html), then run:

```bash
mise trust
make hooks
```

`mise.toml` pins the project's Lefthook version. Git hooks invoke it through mise, so `mise` must be available on `PATH` even when shell activation is disabled.

Pre-commit runs `make check-staged` to check the formatting of staged Go contents, including partially staged files. It preserves staged contents and working files; format and stage corrections explicitly. Pre-push runs `make check` for project-wide linting, vetting, building, and race tests. CI also checks the demo build.

## Running locally

```bash
make build
./lazyjira
```

To test without a Jira account:

```bash
make build-demo
./lazyjira --demo
```

## Nix

If you have Nix with flakes enabled, you can get a complete dev environment
without installing Go or other tools globally:

```bash
nix develop
make check
```

### Adding a Go dependency

After you add a new package to `go.mod`, refresh the Nix lockfile

```bash
make nix-deps
```

You need `gomod2nix` for this. Get it one of two ways

- With Nix `nix develop -c make nix-deps`
- With Go `go install github.com/nix-community/gomod2nix@latest`

Commit `gomod2nix.toml` together with `go.mod` and `go.sum`. If you skip it the `nix` CI job fails with a checksum error

## Code style

- Go standard formatting (`gofmt`)
- Linting via `golangci-lint` (run with `make lint`)
- Keep functions focused and small
- Follow existing patterns in the codebase

## Reporting bugs

Open an issue on GitHub with:
- Steps to reproduce
- Expected vs actual behavior
- Terminal and OS info

## Feature requests

Open an issue describing the use case. Check the roadmap in README first.
