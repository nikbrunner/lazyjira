# Releases

Release with an agent: run [`lj-release`](../.agents/skills/lj-release/SKILL.md). It asks for approval
before it merges the release PR.

This page is the process that skill follows, and every step can also be run by hand.

```mermaid
sequenceDiagram
  participant You
  participant main
  participant PR as Release PR
  participant Actions as GitHub Actions
  You->>main: Push Conventional Commits
  Actions->>PR: release-please opens or updates the release PR with the generated CHANGELOG.md section
  You->>PR: Push the curated Highlights and fresh recordings, and copy the section into the PR description
  You->>main: Merge the release PR
  Actions->>Actions: release-please tags vX.Y.Z and creates the GitHub Release from the PR description
  Actions->>Actions: GoReleaser builds binaries and attaches them
```

## Versioning

[release-please](https://github.com/googleapis/release-please) reads [Conventional Commits](https://www.conventionalcommits.org/)
on `main` and keeps one release PR open that bumps
[`.github/.release-please-manifest.json`](../.github/.release-please-manifest.json). Before `1.0.0`, a breaking change and
`feat` bump the minor version, and `fix` the patch version. A `Release-As: X.Y.Z` footer in a commit body forces the next
version.

## Changelog

release-please writes each release section of `CHANGELOG.md` in the release PR: a heading with the version, compare link,
and date, then the commits grouped by the `changelog-sections` in
[`.github/release-please-config.json`](../.github/release-please-config.json). It inserts the new section above the first
heading that starts with `## [` or `## <digit>`, so every release heading keeps that format.

The curated `### Highlights` go directly under the heading, above the generated groups. release-please publishes the
release PR description as the GitHub Release notes, so the description carries the same section and the release page
matches `CHANGELOG.md`.

## Prepare the release

Every push to `main` makes release-please rebuild the release PR branch and its description, which drops anything added to
them. Push everything for the release first, wait for release-please to finish its run, then curate and merge without
pushing to `main` in between.

Read the version from the release PR title and check out its branch:

```sh
gh pr list --label "autorelease: pending"
gh pr checkout <number>
```

Write the Highlights with
[`lj-changelog`](../.agents/skills/lj-changelog/references/curated-release.md). Then re-record the README screenshot and
demo GIFs so they show the release. This needs [VHS](https://github.com/charmbracelet/vhs) (`brew install vhs`):

```sh
make e2e-update
```

Look at the screenshot and GIFs in `docs/assets/recordings/` before committing them. Then check and push the release
branch:

```sh
make lint-docs
make check
git add CHANGELOG.md docs/assets/recordings/
git commit -m "docs: curate the <version> release notes"
git push
```

The push runs CI on the release PR. Copy the section into the PR description, keeping release-please's header line, the
`---` lines around the notes, and its footer:

```sh
gh pr edit <number> --body-file <file>
```

## Merge the release PR

Merge once CI on the release PR is green:

```sh
gh pr merge <number> --squash
```

The merge starts [`release.yml`](../.github/workflows/release.yml). release-please tags `v<version>` and creates the GitHub
Release with the PR description as its notes. GoReleaser then builds the archives and `checksums.txt` and attaches them,
keeping the notes as they are.

## Verify

```sh
gh run list --workflow release.yml --limit 1
gh release view v<version>
```

The release must list the archives for darwin, linux, and windows, and `checksums.txt`. Its notes must match the version's
`CHANGELOG.md` section.

Install the release through each channel and check the version:

```sh
curl -fsSL https://raw.githubusercontent.com/nikbrunner/lazyjira/main/install.sh | INSTALL_DIR="$(mktemp -d)" sh
mise exec github:nikbrunner/lazyjira@latest -- lazyjira --version
GOBIN="$(mktemp -d)" go install github.com/nikbrunner/lazyjira/cmd/lazyjira@v<version>
```
