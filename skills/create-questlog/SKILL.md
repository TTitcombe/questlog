---
name: create-questlog
description: Use the 'questlog' CLI to build a learning curriculum for the user.
---

Create a learning curriculum for the user's topic ($ARGUMENTS) using the "questlog" CLI.
If no topic was provided by the user, pull out interesting topics from your discussions.

## Questlog
Questlog (https://github.com/TTitcombe/questlog) is a go-based CLI tool that manages reading materials
in learning tracks, useful to compose curricula, TODOs, or just track interest work for later reading.
Check you have questlog installed by running `qlog` in the terminal.
If you don't have it installed, tell the user the have to install it manually by reading the instructions at
https://github.com/TTitcombe/questlog. You should then exit.

### Using questlog
Quick commands for using questlog:
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

Available resource types are: paper · video · book · article · note · idea

**Adding resources: ** `qlog add` has several flags to create entries:
```bash
Usage:
  qlog add [flags]

Examples:
  qlog add --quick "interesting idea about transformers"
  qlog add --title "Attention is All You Need" --type paper --track llm --minutes 45

Flags:
  -h, --help           help for add
      --minutes int    estimated minutes to complete
      --priority int   priority 1 (highest) to 5 (lowest), 0 = unset
  -q, --quick string   quick capture to inbox (no prompts)
      --tags string    comma-separated tags
      --title string   resource title
  -t, --track string   track name (default: inbox) (default "inbox")
      --type string    resource type (paper|video|book|article|note|idea)
      --url string     URL
```

## Create a curriculum
Based on the user's provided topic, you should create a feasible, achievable
curriculum of learning resources that directly work them towards their goal.
This curriculum should be put into an appropriate learning track.
Take into account their current exerience level: if there are foundational works they need to know
before directly tackling their goal, these should also be noted.
Also consider the user's objectives: are they looking for deep technical expertise or to get a high-level,
wide understanding of a topic or field?
A variety of resource types are useful. E.g. unless otherwise stated or necessary, do not overdo academic papers; adding seminal works in the direct topic are encouraged.

1. Check the existing learning tracks in qlog with `qlog track list` to see tracks and `qlog track show <name>` to view resources already in a track. Opt for adding to an existing one if relevant, otherwise create a new one.
2. **Draft the resource list from your own knowledge first.** For well-known papers, videos, and blog posts from established authors and institutions, you already know the canonical URLs — use them directly. Only use web search when you are genuinely unsure whether a resource exists or what its URL is. Cap web searches at 3 total. Sources: arxiv, YouTube, GitHub, well-known technical blogs.
3. If you find resources which are themselves curricula, evaluate the source for technical relevance; use the sub-sources if it seems like a trustworthy source (likes, github stars, expert author etc.).
4. For each resource, note: title, URL, type, estimated minutes, and priority (1=foundational, 5=niche/advanced).
5. Aim for 15–25 resources covering the topic. More is not better — cut the least relevant until you're in range.
6. Order by priority: foundational/introductory items first (priority 1–2), niche/advanced last (priority 4–5).
7. Add all resources to the qlog track in a **single batched bash call** (chain multiple `qlog add` commands with `&&` or `;` in one Bash tool invocation per thematic group). Do not make one tool call per resource.
