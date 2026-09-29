#!/bin/sh
# Installs the lazyjira binary from GitHub Releases.
#
#   curl -fsSL https://raw.githubusercontent.com/nikbrunner/lazyjira/main/install.sh | sh
#
# Environment:
#   VERSION      release tag to install, e.g. v0.6.5 (default: latest release)
#   INSTALL_DIR  target directory (default: $HOME/.local/bin)
#   LAZYJIRA_RELEASES_URL  release download base URL (default: GitHub Releases)

set -eu

releases="${LAZYJIRA_RELEASES_URL:-https://github.com/nikbrunner/lazyjira/releases}"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
version="${VERSION:-}"

fail() {
	echo "lazyjira install: $*" >&2
	exit 1
}

case "$(uname -s)" in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fail "unsupported OS $(uname -s); on Windows use go install or mise" ;;
esac

case "$(uname -m)" in
x86_64 | amd64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fail "unsupported architecture $(uname -m)" ;;
esac

command -v curl >/dev/null 2>&1 || fail "curl is required"
if command -v sha256sum >/dev/null 2>&1; then
	sha256() { sha256sum "$1" | cut -d ' ' -f 1; }
elif command -v shasum >/dev/null 2>&1; then
	sha256() { shasum -a 256 "$1" | cut -d ' ' -f 1; }
else
	fail "sha256sum or shasum is required"
fi

if [ -n "$version" ]; then
	download="$releases/download/$version"
else
	download="$releases/latest/download"
fi
archive="lazyjira_${os}_${arch}.tar.gz"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $archive from $download"
curl -fsSL -o "$tmp/$archive" "$download/$archive" || fail "could not download $download/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$download/checksums.txt" || fail "could not download $download/checksums.txt"

expected="$(awk -v name="$archive" '$2 == name { print $1 }' "$tmp/checksums.txt")"
[ -n "$expected" ] || fail "checksums.txt has no entry for $archive"
[ "$(sha256 "$tmp/$archive")" = "$expected" ] || fail "checksum mismatch for $archive"

tar -xzf "$tmp/$archive" -C "$tmp" lazyjira
mkdir -p "$install_dir"
install -m 0755 "$tmp/lazyjira" "$install_dir/lazyjira"

echo "Installed $("$install_dir/lazyjira" --version) to $install_dir/lazyjira"
case ":$PATH:" in
*":$install_dir:"*) ;;
*) echo "Add $install_dir to your PATH to run lazyjira." ;;
esac
