# Releases

Release with an agent: run [`lj-release`](../.agents/skills/lj-release/SKILL.md). It asks for approval
before it merges the release PR.

This page is the process that skill follows, and every step can also be run by hand.

```mermaid
sequenceDiagram
  participant You
  participant main
  participant Actions as GitHub Actions
  You->>main: Push Conventional Commits
  Actions->>main: release-please opens or updates the release PR
  You->>main: Push the curated CHANGELOG.md section and fresh recordings
  You->>Actions: Merge the release PR
  Actions->>Actions: release-please tags vX.Y.Z and creates the GitHub Release
  Actions->>Actions: GoReleaser builds binaries and replaces the notes with the curated section
```

## Versioning

[release-please](https://github.com/googleapis/release-please) reads [Conventional Commits](https://www.conventionalcommits.org/)
on `main` and keeps one release PR open that bumps
[`.github/.release-please-manifest.json`](../.github/.release-please-manifest.json). Before `1.0.0`, `feat` bumps the minor
version and `fix` the patch version. A `Release-As: X.Y.Z` footer in a commit body forces the next version.

`CHANGELOG.md` is written by hand. release-please never edits it (`skip-changelog` in
[`.github/release-please-config.json`](../.github/release-please-config.json)).

## Prepare the release commit

Read the proposed version from the release PR title. Curate its `CHANGELOG.md` section with
[`lj-changelog`](../.agents/skills/lj-changelog/references/curated-release.md) in release mode:
highlights by impact, the trimmed log below them, and the heading date set to the release day. The version in the heading
must match the release PR.

Re-record the README GIFs so they show the release. This needs [VHS](https://github.com/charmbracelet/vhs)
(`brew install vhs`):

```sh
make e2e-update
```

Look at the GIFs in `docs/assets/recordings/` before committing them. Then check and push everything to
`main`:

```sh
make lint-docs
make check
git add CHANGELOG.md docs/assets/recordings/
git commit -m "docs: prepare the <version> release"
git push origin main
```

release-please force-pushes its PR branch on every run, so commits added to that branch are lost. The curated section and the
recordings live on `main`.

## Merge the release PR

release-please opens its PR with `GITHUB_TOKEN`, so CI does not run on it. The curated commit's CI run on `main` covers the
release; merge once it is green:

```sh
gh pr list --label "autorelease: pending"
gh pr merge <number> --squash
```

The merge starts [`release.yml`](../.github/workflows/release.yml). release-please tags `v<version>` and creates the GitHub
Release. GoReleaser then builds the archives and `checksums.txt`, attaches them, and replaces the release notes with the
`CHANGELOG.md` section for that version. The job fails when the section is missing.

## Verify

```sh
gh run list --workflow release.yml --limit 1
gh release view v<version>
```

The release must list the archives for darwin, linux, and windows, and its notes must match the curated section. The release
must also list `checksums.txt`.

Install the release through each channel and check the version:

```sh
curl -fsSL https://raw.githubusercontent.com/nikbrunner/lazyjira/main/install.sh | INSTALL_DIR="$(mktemp -d)" sh
mise exec github:nikbrunner/lazyjira@latest -- lazyjira --version
GOBIN="$(mktemp -d)" go install github.com/nikbrunner/lazyjira/cmd/lazyjira@v<version>
```
