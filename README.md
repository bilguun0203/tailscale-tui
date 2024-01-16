# Tailscale TUI

Terminal based Tailscale status viewer written in [Golang](https://go.dev/) with help of [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

## Install

```sh
go install github.com/bilguun0203/tailscale-tui@latest
```

## Usage

```sh
tailscale-tui
```

### Shortcuts

- `↑/k` `↓/j` - up/down
- `→/l/pgdn` `←/h/pgup` - next/prev page
- `g/home` `G/end` - go to start/end
- `q` `Ctrl+c` - quit
- `/` - filter
- `y` - copy ipv4 of the selected node
- `?` - expand/collapse help
