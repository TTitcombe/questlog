# questlog

!! WIP !!

Your personal learning tracker.

Capture ideas and resources, organise them into tracks (skill trees), track your progress, and get focused session suggestions when you have spare time.

Data lives in `~/.questlog/` as plain Markdown files + a JSON index, making it easy to read and edit directly, including by agents like Claude Code.

## Getting started
### Install

```bash
make install
# or
go install ./cmd/qlog
```

### Agent skill

Questlog is designed to be used with an agent (e.g. Claude Code)
doing the hard work to find and prioritise resources.
The repository contains a SKILL teaching your agent how to do this.\
Install the skill in your agent with:
```bash
npx skills add TTitcombe/questlog
```

And run it with a description of what you want to learn about:

```bash
/create-questlog Understand LLM architecture and training, so that I can create tailored models for my team.
```

If you don't provide a topic when invoking the skill,
your agent will default to inferring topics of interest to you from your conversations.

### Accessing a track

With resources in a learning track, enter "focus" mode to the next queued resources
you can achieve without the time you have available:

```bash
# Minutes you have available to focus (defaults to 30).
qlog focus --track <track-name> --minutes 60
```

This will open the resources in order. You can update their status and add notes to each one.
(WORK TODO).

## Quickstart commands

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

## Detail
### Commands

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

### Resource types

`paper` · `video` · `book` · `article` · `note` · `idea`

### Data format

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

## License

MIT

See [LICENSE](./LICENSE).
