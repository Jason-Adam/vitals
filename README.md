# vitals

A terminal statusline for [Claude Code](https://docs.anthropic.com/en/docs/claude-code) sessions.

![vitals demo](demo.gif)

## Install

```bash
go install github.com/Jason-Adam/vitals/cmd/vitals@latest
```

## Setup

Add to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "vitals"
  }
}
```

To customize, run `vitals --init` to generate a config at `~/.config/vitals/config.toml`.

## License

[MIT](LICENSE)
