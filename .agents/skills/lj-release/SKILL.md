---
name: lj-release
description:
  "Prepare and ship a lazyjira release. Use this when cutting a version, preparing release notes, or merging the
  release-please PR. Ask for approval before merging the release PR."
---

# Release lazyjira

Follow [`docs/releases.md`](../../../docs/releases.md). It is the maintainer source of truth for the release PR, the curated
release section, and checking the GitHub Release.

Read the version from the release PR title after release-please has finished its run for the latest push to `main`; each
run can change the version. Curate the section on the release branch with [`lj-changelog`](../lj-changelog/SKILL.md).

Re-record the screenshot and GIFs with `make e2e-update` and show Nik the new `docs/assets/recordings/screenshot.png` before
committing them on the release branch with the curated section.

Nothing lands on `main` between curating and merging. Ask for explicit approval immediately before merging the release PR.
