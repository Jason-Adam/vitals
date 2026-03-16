# tail-claude-hud

A terminal statusline for Claude Code sessions. Built with Go.

Reads JSON from Claude Code's stdin pipe on every tick, parses transcript state, gathers git/env data, and renders a styled multi-line statusline. The full cycle completes in single-digit milliseconds.

## Requirements

- Go 1.25+
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) with statusline support

## Install

```bash
go install github.com/kylesnowschwartz/tail-claude-hud@latest
```

Or build from source:

```bash
git clone git@github.com:kylesnowschwartz/tail-claude-hud.git
cd tail-claude-hud
go build -o tail-claude-hud ./cmd/tail-claude-hud
```

## Setup

Generate a default config:

```bash
tail-claude-hud --init
```

This creates `~/.config/tail-claude-hud/config.toml`. Then point Claude Code's statusline at the binary.

## Configuration

Layout is TOML. Each `[[line]]` defines a row of widgets:

```toml
[[line]]
widgets = ["model", "context", "duration", "session"]

[[line]]
widgets = ["agents"]

[[line]]
widgets = ["tools"]
```

Available widgets: `model`, `context`, `directory`, `git`, `project`, `env`, `duration`, `tools`, `agents`, `todos`, `session`, `thinking`.

Each widget is a pure function that returns a styled string or `""` when it has nothing to show. Configure only what you want to see.

### CLI flags

```
tail-claude-hud [flags]
  --init           generate a default config file
  --dump-current   render from a transcript file instead of stdin
```

## Development

Requires [just](https://github.com/casey/just) for task running.

```bash
just              # run tests
just build        # go build
just test         # go test ./... -count=1
just test-race    # race detector
just bench        # benchmarks
just check        # fmt + vet + test
just dump         # build + render from current session
just run-sample   # pipe testdata through the binary
```

## Related

- [tail-claude](https://github.com/kylesnowschwartz/tail-claude) — Terminal TUI for reading Claude Code session logs

## License

[MIT](LICENSE)
