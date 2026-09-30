---
name: lj-release
description:
  "Prepare and ship a lazyjira release. Use this when cutting a version, preparing release notes, or merging the
  release-please PR. Ask for approval before merging the release PR."
argument-hint: "[version, for example 0.7.0; defaults to the release PR's version]"
allowed-tools: Bash Read Edit
---

# Release lazyjira

Follow [`docs/releases.md`](../../../docs/releases.md). It is the maintainer source of truth for the release PR, the curated
changelog commit, and checking the GitHub Release.

Prepare the release notes with [`lj-changelog`](../lj-changelog/SKILL.md) in release mode before merging:
it curates the section and sets the release date. The section's version must match the release PR title.

Re-record the screenshot and GIFs with `make e2e-update` and show Nik the new `docs/assets/recordings/screenshot.png` before committing
them with the changelog.

Ask for explicit approval immediately before merging the release PR.
