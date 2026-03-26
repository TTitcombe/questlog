# Agent instructions for questlog

This repo is the source for **questlog** (`qlog`), the user's personal learning tracker.
The installed binary uses data at `~/.questlog/` as plain Markdown files.

## When to use qlog proactively

- **Capture to inbox** — if you encounter a useful resource (paper, video, article, tool) during a conversation that the user hasn't explicitly saved, add it: `qlog add --quick "<title or description>"`
- **Add with full details** — if you have a URL and type: `qlog add --title "..." --type paper --url "..." --track <track> --minutes <est>`
- **Check context** — at the start of a conversation about a topic, check if the user has relevant resources: `qlog search "<topic>"` or `qlog track show <track>`
- **Update progress** — if the user mentions they've read or finished something: `qlog done <id>` or `qlog progress <id> <0-100>`

## Key commands

```bash
qlog add --quick "idea or resource"                          # instant inbox capture
qlog add --title "..." --type paper --track llm --minutes 45 --priority 1
qlog inbox                                                   # view uncategorised items
qlog classify <id> --track <name>                            # move inbox item to track
qlog track list                                              # all tracks + progress
qlog track show <name>                                       # resources in a track
qlog guide --track <name>                                    # prioritized reading guide
qlog focus --track <name> --minutes 30                       # session suggestions
qlog search "<query>"                                        # search by title/tags
qlog status                                                  # overall progress overview
qlog done <id>                                               # mark resource complete
qlog progress <id> <0-100>                                   # set progress %
```

## Resource types
`paper` · `video` · `book` · `article` · `note` · `idea`

## Priority
1 = highest, 5 = lowest, 0 = unset. Use priority when adding resources an agent judges to be especially important for the user's current goals.

## IDs
Each resource has a slug ID (e.g. `attention-is-all-you-need`) shown in dim text in all list/show output. Use this ID in `done`, `progress`, `classify` commands.

## Direct file editing
Resource files at `~/.questlog/tracks/<track>/resources/<id>.md` can be read and edited directly. Run `qlog index rebuild` after any direct edits to resync the search index.
