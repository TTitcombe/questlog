# questlog

!! WIP !!

Your personal learning tracker.

Capture ideas and resources, organise them into tracks (skill trees), track your progress, and get focused session suggestions when you have spare time.

Data lives in `~/.questlog/` as plain Markdown files + a JSON index, making it easy to read and edit directly, including by agents like Claude Code.

## Install

```bash
make install
# or
go install ./cmd/qlog
```

## Quick start

```bash
# Capture a quick idea to your inbox
qlog add --quick "read about attention mechanisms"

# Create a learning track
qlog track new llm --description "LLM fundamentals"

# Add a resource with full details
qlog add --title "Attention is All You Need" --type paper --track llm --minutes 45

# View your inbox
qlog inbox

# Move an inbox item to a track
qlog classify <id> --track llm

# Get a 30-minute focus session
qlog focus --track llm --minutes 30

# Mark something done
qlog done <id>

# Check overall progress
qlog status
```

## Commands

| Command | Description |
|---|---|
| `qlog add` | Interactive prompt to add a resource |
| `qlog add --quick "text"` | Instant inbox capture |
| `qlog inbox` | View uncategorised inbox items |
| `qlog track new <name>` | Create a new track |
| `qlog track list` | List all tracks with progress |
| `qlog track show <name>` | Detailed track view |
| `qlog list` | List resources (filter with `--track`, `--status`, `--type`) |
| `qlog done <id>` | Mark a resource as done |
| `qlog progress <id> <0-100>` | Set progress percentage |
| `qlog classify <id>` | Move an inbox item to a track |
| `qlog focus` | Suggest a focus session (`--track`, `--minutes`) |
| `qlog search <query>` | Search by title and tags |
| `qlog status` | Overview of all tracks and progress |

## Resource types

`paper` · `video` · `book` · `article` · `note` · `idea`

## Data format

Each resource is a Markdown file with YAML frontmatter:

```markdown
---
title: "Attention is All You Need"
type: paper
url: https://arxiv.org/abs/1706.03762
tags: [transformers, attention, llm]
track: llm
added: 2026-03-22
estimated_minutes: 45
status: unread
progress: 0
---

Your personal notes here...
```

Tracks live at `~/.questlog/tracks/<name>/`, inbox items at `~/.questlog/inbox/`.

If you edit files directly (or an agent does), run `qlog index rebuild` to resync the search index.
