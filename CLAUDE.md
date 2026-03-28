# CLAUDE.md

## What This Is

vitals is a Go binary that renders a terminal statusline for Claude Code sessions. Claude Code pipes JSON to stdin on every tick; this binary parses it, gathers supplementary data (transcript state, git status), and prints a styled multi-line statusline to stdout. The entire cycle must complete in single-digit milliseconds because it runs on every keypress/tick.

## Build & Test

Uses `make` for all tasks:

```sh
make              # default: run tests
make build        # go install ./cmd/vitals
make test         # go test ./... -count=1
make test-race    # go test -race ./... -count=1
make bench        # go test -bench=. -benchmem ./internal/... -count=1
make check        # fmt + vet + test
make dump         # build + render from current session's transcript
make run-sample   # pipe testdata/sample-stdin.json through the binary
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

2. **gather** (`internal/gather`): Inspects which widgets are active in the config, spawns goroutines *only* for the data sources those widgets need (transcript, git). A `sync.WaitGroup` gates the render stage.

3. **render** (`internal/render`): Walks configured lines, calls each widget's `RenderFunc` from the registry, joins non-empty results with the separator, and ANSI-truncates to terminal width.

4. **widget** (`internal/render/widget`): 18 registered widgets (model, context, cost, directory, git, project, service, duration, tools, agents, todos, tokens, lines, messages, speed, permission, usage, worktree). Each is a pure function: `(RenderContext, Config) -> WidgetResult`. Returns empty when it has nothing to show.

## Key Design Decisions

**Fail-open config**: `config.LoadHud()` never returns nil or an error. Missing or corrupt TOML yields defaults. The statusline must always render something.

**Incremental transcript reads**: `transcript.StateManager` tracks byte offsets per transcript path (keyed by SHA-256 hash). Each tick reads only new bytes (O(delta) not O(n)). Extraction state is snapshotted to disk so the full tool/agent/todo history survives process restarts.

**Never write to stderr**: Claude Code owns the terminal. Any stderr output corrupts the display. Debug logging goes to `~/.claude/plugins/vitals/debug.log` and is gated behind `VITALS_DEBUG=1`.

**Conditional goroutines**: The gather stage checks which widgets are configured before spawning work. If no transcript widgets are active, no transcript parsing runs.

**Hook-based permission detection**: The binary doubles as a Claude Code hook handler via `vitals hook <event>`. The `PermissionRequest` hook writes a breadcrumb file to `~/.config/vitals/waiting/{session_id}`; `PostToolUse` and `Stop` hooks remove it. The statusline gather stage scans this directory (skipping its own session) to detect other sessions blocked on permission approval. Breadcrumbs older than 120s are ignored (covers hard crashes).

## Transcript Processing (Three Layers)

- **transcript.go** — Parses individual JSONL entries and classifies content blocks (tool_use, tool_result, thinking, text). Handles sidechain filtering (sub-agent user messages are excluded).
- **extractor.go** — Stateful processor that accumulates tools, agents, and todos across entries. Handles agent lifecycle (launch, async results, task notifications), todo mutations (TodoWrite replaces all, TaskCreate/TaskUpdate mutate), and the scrolling divider counter.
- **state.go** — Byte-offset persistence for incremental reads. Embeds the extraction snapshot so state survives across ticks.

## Config

TOML at `~/.config/vitals/config.toml` (or legacy `~/.claude/plugins/vitals/config.toml`). Generate defaults with `vitals --init`.

Layout is configured as `[[line]]` arrays with widget name lists. Default is two lines: summary and agents.

## Stdin JSON Contract

Claude Code pipes JSON to stdin on every tick. Canonical reference: https://code.claude.com/docs/en/statusline#available-data

Key fields: `model.*`, `session_id`, `transcript_path`, `cost.*`, `context_window.*`, `worktree.*`. The `context_window.current_usage` fields are per-call snapshots (not session totals). Session-level cumulative totals are in `total_input_tokens` and `total_output_tokens`.
