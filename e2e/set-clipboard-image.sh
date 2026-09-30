#!/usr/bin/env bash
# Puts a PNG on the system clipboard so a tape can paste it with ctrl+v.
set -euo pipefail
file="$(cd "$(dirname "$1")" && pwd)/$(basename "$1")"
if command -v osascript >/dev/null; then
    osascript -e "set the clipboard to (read (POSIX file \"$file\") as «class PNGf»)"
elif [[ -n "${WAYLAND_DISPLAY:-}" ]]; then
    wl-copy --type image/png <"$file"
else
    xclip -selection clipboard -t image/png -i "$file"
fi
