# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & install

```bash
go build ./...          # compile check
make build              # outputs bin/qlog
make install            # installs to ~/go/bin (requires ~/go/bin on PATH)
```

There are no tests yet. The binary entry point is `cmd/qlog/main.go`.

## Architecture

This is a Go CLI tool (`qlog`) for personal learning tracking. Data is stored as Markdown files + a JSON index at `~/.questlog/`.

### Layer structure

```
cmd/qlog/          → binary entry point, wires root cobra command
internal/model/    → pure data types (no dependencies)
internal/store/    → filesystem I/O; implements the Store interface
internal/cli/      → cobra command handlers; depend on Store interface
internal/cli/ui/   → lipgloss styles and rendering helpers
```

### Data flow

All CLI commands receive a `Store` interface (initialised once in `root.go`'s `PersistentPreRunE`). The concrete implementation is `FSStore` in `internal/store/fs.go`.

Every write operation (`SaveResource`, `DeleteResource`, `MoveToTrack`) atomically updates `~/.questlog/index.json` via a write-to-tmp-then-rename pattern. If the index is missing on startup, `FSStore.New()` auto-rebuilds it by walking all `.md` files.

### Data format

Each resource is a Markdown file with YAML frontmatter. The `frontmatter` struct in `store/markdown.go` is the authoritative definition of what gets persisted. The `model.Resource` struct adds runtime-only fields (`ID`, `FilePath`, `Notes`) that are never written to the frontmatter.

- Track resources: `~/.questlog/tracks/<track>/resources/<slug>.md`
- Inbox items: `~/.questlog/inbox/YYYY-MM-DD-<slug>.md`
- Track metadata: `~/.questlog/tracks/<track>/track.json`
- Search index: `~/.questlog/index.json` (frontmatter fields only, no body text)

### Key design decisions

- **Slug = ID**: resource identity is derived from the title slug (see `store/slug.go`), making files human-readable and directly editable by agents. No UUIDs.
- **Store interface over FSStore**: CLI handlers always take `func() *store.FSStore` — swap in a different implementation for testing.
- **Index excludes body text**: `index.json` only stores frontmatter fields for fast in-memory search. Full-text search of notes requires reading individual files.
- **`priority` field**: integer 1–5 (1 = highest, 0 = unset). Used by `qlog guide` to rank the reading order. Unset priorities sort last.

### Adding a new command

1. Create `internal/cli/<name>.go` with a `newXxxCmd(getStore func() *store.FSStore) *cobra.Command` function.
2. Register it in `root.go`'s `root.AddCommand(...)` list.
