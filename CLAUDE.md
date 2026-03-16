# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

tail-claude-hud is a Go binary that renders a terminal statusline for Claude Code sessions. Claude Code pipes JSON to stdin on every tick; this binary parses it, gathers supplementary data (transcript state, git status, environment counts), and prints a styled multi-line statusline to stdout. The entire cycle must complete in single-digit milliseconds because it runs on every keypress/tick.

It combines techniques from two reference projects in `.cloned-sources/`: `tail-claude` (Go transcript parsing) and `claude-hud` (TypeScript statusline plugin with HUD-first design).

## Build & Test

Uses `just` (justfile) for all tasks:

```sh
just              # default: run tests
just build        # go build -o bin/tail-claude-hud ./cmd/tail-claude-hud
just test         # go test ./... -count=1
just test-race    # go test -race ./... -count=1
just bench        # go test -bench=. -benchmem ./internal/... -count=1
just check        # fmt + vet + test
just dump         # build + render from current session's transcript
just run-sample   # pipe testdata/sample-stdin.json through the binary
```

Run a single test:
```sh
go test ./internal/transcript/ -run TestExtractContentBlocks -count=1
```

## Architecture: The Four-Stage Pipeline

Every invocation follows a strict linear pipeline. Each stage is a separate package with no backward dependencies.

```
stdin → gather → render → stdout
```

1. **stdin** (`internal/stdin`): Decodes JSON from Claude Code, computes context percentage, persists a snapshot to disk for `--dump-current` mode.

2. **gather** (`internal/gather`): Inspects which widgets are active in the config, spawns goroutines *only* for the data sources those widgets need (transcript, git, env). A `sync.WaitGroup` gates the render stage.

3. **render** (`internal/render`): Walks configured lines, calls each widget's `RenderFunc` from the registry, joins non-empty results with the separator, and ANSI-truncates to terminal width.

4. **widget** (`internal/render/widget`): 12 registered widgets (model, context, directory, git, project, env, duration, tools, agents, todos, session, thinking). Each is a pure function: `(RenderContext, Config) -> string`. Returns `""` when it has nothing to show.

## Key Design Decisions

**Fail-open config**: `config.LoadHud()` never returns nil or an error. Missing or corrupt TOML yields defaults. The statusline must always render something.

**Incremental transcript reads**: `transcript.StateManager` tracks byte offsets per transcript path (keyed by SHA-256 hash). Each tick reads only new bytes (O(delta) not O(n)). Extraction state is snapshotted to disk so the full tool/agent/todo history survives process restarts.

**Never write to stderr**: Claude Code owns the terminal. Any stderr output corrupts the display. Debug logging goes to `~/.claude/plugins/tail-claude-hud/debug.log` and is gated behind `TAIL_CLAUDE_HUD_DEBUG=1`.

**Conditional goroutines**: The gather stage checks which widgets are configured before spawning work. If no transcript widgets are active, no transcript parsing runs.

## Transcript Processing (Three Layers)

The transcript package has three distinct responsibilities:

- **transcript.go** — Parses individual JSONL entries and classifies content blocks (tool_use, tool_result, thinking, text). Handles sidechain filtering (sub-agent user messages are excluded).
- **extractor.go** — Stateful processor that accumulates tools, agents, and todos across entries. Handles agent lifecycle (launch, async results, task notifications), todo mutations (TodoWrite replaces all, TaskCreate/TaskUpdate mutate), and the scrolling divider counter.
- **state.go** — Byte-offset persistence for incremental reads. Embeds the extraction snapshot so state survives across ticks.

## Config

TOML at `~/.config/tail-claude-hud/config.toml` (or legacy `~/.claude/plugins/tail-claude-hud/config.toml`). Generate defaults with `tail-claude-hud --init`.

Layout is configured as `[[line]]` arrays with widget name lists. Default is three lines: summary, agents, tools.

## Reference Projects

`.cloned-sources/claude-hud/` — Original TypeScript plugin. Reference for UI patterns, color choices, widget behavior.
`.cloned-sources/tail-claude/` — Go predecessor. Reference for transcript entry schema, content block parsing, tool categorization.
