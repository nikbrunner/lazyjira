# lazyjira

<p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/nikbrunner/lazyjira" alt="Go"></a>
  <a href="https://github.com/nikbrunner/lazyjira/releases"><img src="https://img.shields.io/github/v/release/nikbrunner/lazyjira" alt="Release"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  <a href="https://github.com/nikbrunner/lazyjira/releases"><img src="https://img.shields.io/github/downloads/nikbrunner/lazyjira/total?label=downloads" alt="Downloads"></a>
</p>

<p align="center">
  <img src="assets/lazygit.png" width="100%" alt="lazyjira banner">
</p>

Terminal UI for Jira. Like [lazygit](https://github.com/jesseduffield/lazygit) but for Jira.

Based on the [original project](https://github.com/textfuel/lazyjira), created by textfuel and its contributors. The original MIT license and copyright notice are preserved in [LICENSE](LICENSE).

Changing a status in Jira's web UI takes multiple clicks, and pages can take seconds to load. lazyjira is a keyboard-driven terminal UI for browsing issues, updating statuses, and reading descriptions.

<p>
  <img src="e2e/golden/00_preview.gif" width="67%" alt="preview">&nbsp;<img src="e2e/golden/00_preview_vertical.gif" width="31%" alt="preview vertical">
</p>

### Demo mode

Try without a Jira account (build from source required):

```
make build-demo
./lazyjira --demo
```

## Features

- **JQL search** with autocomplete, syntax highlighting, and persistent history
- **4-panel layout** - issues, projects, detail, status - with vim-style navigation
- **Inline editing** - transitions, priority, assignee, labels, comments, description (`$EDITOR`)
- **Configurable** - custom keybindings (including navigation keys), JQL tabs, issue columns, custom fields
- **Themes** - default ANSI palette plus all four Catppuccin flavors (Latte, Frappé, Macchiato, Mocha)
- **Adaptive** - side-by-side or stacked layout, mouse support

## Installation

Requires Go 1.25.8 or later.

Install from the main branch:

```sh
go install github.com/nikbrunner/lazyjira/v2/cmd/lazyjira@main
```

Go installs `lazyjira` in `GOBIN` if it is set, or in the `bin` folder under `GOPATH` otherwise. Your `PATH` tells the terminal where to find commands.

On macOS or Linux using Zsh add these lines at the end:

```sh
go_bin="$(go env GOBIN)"
if [ -z "$go_bin" ]; then
  go_bin="$(go env GOPATH)/bin"
fi
export PATH="$PATH:$go_bin"
```

```sh
lazyjira --version
```

On Windows, run `go env GOBIN` in PowerShell. If it prints nothing, run `go env GOPATH` and add `\bin` to the end of that path. Press `Win+R`, enter `sysdm.cpl`, then open Advanced → Environment Variables → your user `Path` → Edit → New. Add the directory, open a new PowerShell window, and run `lazyjira --version`.

Or build from source:

```sh
git clone https://github.com/nikbrunner/lazyjira.git
cd lazyjira
make build
./lazyjira
```

## Setup

Run `lazyjira`. On first launch the setup wizard asks for your Jira type (Cloud or Server/Data Center), host, and credentials.

### Jira Cloud

Provide your email and an API token (also called Personal Access Token / PAT).

Create one at <https://id.atlassian.com/manage-profile/security/api-tokens>

### Jira Server / Data Center

Provide your Personal Access Token (PAT). No email needed.

Generate a PAT in Jira: Profile > Personal Access Tokens > Create token.

For environments that require client certificates (mTLS), see [Configuration](docs/Config.md#tls).

Credentials saved to `~/.config/lazyjira/auth.json`.

## Themes

lazyjira supports the default ANSI 16 palette, Catppuccin presets, auto light/dark selection, and custom palette overrides. Set `gui.theme` in `~/.config/lazyjira/config.yml`:

```yaml
gui:
  theme: catppuccin-mocha
```

See the [Themes docs](docs/Config.md#themes) for supported values, palette overrides, and examples.

## Usage

```
lazyjira                 # start
lazyjira auth            # re-authenticate
lazyjira logout          # clear credentials
lazyjira --dry-run       # read-only mode (no writes to Jira)
lazyjira --log app.log   # log API requests to file
lazyjira --version       # show version
```

Press `?` inside the app for all keybindings.

## Documentation

- [Configuration](docs/Config.md) - config file, keybindings, issue tabs, custom fields, git integration
- [Keybindings](docs/Keybindings.md) - full list of default keys
- [Custom Fields](docs/Custom_Fields.md) - displaying Jira custom fields

## License

MIT
