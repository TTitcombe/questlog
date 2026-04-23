# create-questlog Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the `create-questlog` skill with a goal-builder skill that conducts a guided interview, researches real resources via web search, and executes qlog CLI commands to set up a fully structured goal with tracks, milestones, and a starter curriculum.

**Architecture:** Single `SKILL.md` file replacing the existing one at `~/.claude/skills/create-questlog/SKILL.md`. Four phases: goal interview → curriculum research → plan presentation → qlog execution. No changes to the qlog binary.

**Tech Stack:** Claude Code skills (Markdown), qlog CLI, WebSearch tool

---

## Exact qlog CLI interface (verified from source)

```bash
# Goals
qlog goal new "Goal Title"              # title is positional arg; --description/-d optional
qlog goal list
qlog goal show <slug>
qlog goal milestone add <goal-slug>     # --description/-d "..." --deadline YYYY-MM-DD --artifact <id>
qlog goal milestone done <goal-slug> <milestone-id>

# Tracks
qlog track new <name>                   # --description/-d "..." --goal <goal-slug>
qlog track list
qlog track show <name>
qlog track set-goal <track> <goal-slug>
qlog track depends-on <track> <prerequisite-track>
qlog track milestone add <track-name>   # --description/-d "..." --deadline YYYY-MM-DD --artifact <id>
qlog track milestone done <track-name> <milestone-id>

# Resources
qlog add --title "..." --type <type> --track <name> --minutes <n> --url "..." --priority <1-5>
# types: paper · video · book · article · note · idea
```

---

### Task 1: Write the complete SKILL.md

**Files:**
- Overwrite: `/Users/tom/.claude/skills/create-questlog/SKILL.md`

- [ ] **Step 1: Write the full SKILL.md**

Write the following content to `/Users/tom/.claude/skills/create-questlog/SKILL.md`:

```markdown
---
name: create-questlog
description: Build a structured learning goal with tracks, milestones, and a researched curriculum using the questlog CLI.
---

Build a structured learning goal for the user using the questlog CLI.
Follow the four phases below: goal interview → curriculum research → plan presentation → execution.

## Prerequisites

Check questlog is installed:
```bash
qlog --version
```
If not installed, tell the user to follow the instructions at https://github.com/TTitcombe/questlog and exit.

## qlog CLI Reference

```bash
# Goals — title is a positional argument; slug is auto-derived from it
qlog goal new "Goal Title"              # --description/-d "optional description"
qlog goal list
qlog goal show <slug>
qlog goal milestone add <goal-slug>     # --description/-d "..." --deadline YYYY-MM-DD
qlog goal milestone done <goal-slug> <milestone-id>

# Tracks
qlog track new <name>                   # --description/-d "..." --goal <goal-slug>
qlog track list
qlog track show <name>
qlog track set-goal <track> <goal-slug>
qlog track depends-on <track> <prerequisite-track>
qlog track milestone add <track-name>   # --description/-d "..." --deadline YYYY-MM-DD
qlog track milestone done <track-name> <milestone-id>

# Resources
qlog add --title "..." --type <type> --track <name> --minutes <n> --url "..." --priority <1-5>
# types: paper · video · book · article · note · idea
```

---

## Phase 1: Goal Interview

Conduct a free-form conversation to understand the user's goal. Ask one question at a time. Do not move to research until you have extracted all four of the following:

| What to extract | Notes |
|-----------------|-------|
| **Goal statement** | What the user wants to be able to do / be known for |
| **Time horizon** | Near-term checkpoint (3–6 months) + full arc (1–2 years) |
| **Current baseline** | What's already strong, what's rusty, what's missing entirely |
| **Domain(s)** | The technical areas the goal spans — drives source pack selection |

After the conversation, synthesise into a structured summary and ask the user to confirm before proceeding:

```
Goal: [title] · [timeline]
Pillars:
  1. [Pillar Name]  (domain: ai-ml)
  2. [Pillar Name]  (domain: systems)
  3. [Pillar Name]  (domain: startup-product)
Domain packs activated: ai-ml, systems, startup-product
```

**Capability pillars** are 3–5 named areas that together constitute the goal. Each pillar becomes one track in qlog. The baseline shapes the curriculum — someone strong on fundamentals but weak on deployment gets different resources than someone starting from scratch.

Do not proceed to Phase 2 until the user confirms the summary.

---

## Phase 2: Curriculum Research

For each capability pillar, run targeted web searches to find 5–10 real, linkable resources. Use the source packs below based on domains detected in Phase 1. Multiple packs can be active simultaneously.

Searches must be specific:
- ✓ "best course production LLM deployment 2025"
- ✓ "Papers With Code transformer inference optimization 2024"
- ✗ "AI resources" (too generic)

For each resource, note: title, URL, type, estimated minutes, and which track it belongs to.

For well-known canonical works (e.g. "Attention is All You Need", fast.ai Practical Deep Learning), use the URL directly without searching. For frontier or recent resources, always search — training data may be stale.

### Source Packs

**General Tech** (always active):
- Hacker News (news.ycombinator.com)
- Engineering blogs: Stripe, Cloudflare, Netflix, Uber tech blogs
- MIT OpenCourseWare (ocw.mit.edu), Stanford Online (online.stanford.edu)
- Official documentation

**AI/ML**:
- Hugging Face Daily Papers (huggingface.co/papers)
- Papers With Code (paperswithcode.com)
- fast.ai (fast.ai)
- DeepLearning.ai (deeplearning.ai)
- Lab blogs: Anthropic, OpenAI, DeepMind, Meta AI Research

**Systems/Infrastructure**:
- USENIX (usenix.org)
- High Scalability (highscalability.com)
- Brendan Gregg (brendangregg.com)
- ACM Digital Library (dl.acm.org)

**Startup/Product**:
- First Round Review (review.firstround.com)
- Lenny's Newsletter (lennysnewsletter.com)
- Paul Graham essays (paulgraham.com)

**Web/Frontend**:
- MDN Web Docs (developer.mozilla.org)
- Framework official blogs (React, Vue, Svelte, etc.)
- web.dev

**Courses & Video**:
- YouTube (domain-specific channels: Karpathy, Yannic Kilcher, 3Blue1Brown for AI; Fireship for general SWE; CppCon for systems)
- Coursera (coursera.org), edX (edx.org)
- Conference recordings on YouTube (NeurIPS, ICML for AI; USENIX, Strange Loop for systems)

**Books**:
- O'Reilly (oreilly.com) — primary for tech
- Manning (manning.com)
- Pragmatic Programmers (pragprog.com)
- MIT Press (mitpress.mit.edu)

### Resource Quality Bar

- Target 5–10 resources per track; more is not better — cut the least relevant
- Mix types: don't fill a track with only papers or only videos
- Prefer resources with a specific URL, title, and author you can verify

---

## Phase 3: Plan Presentation

Present the full plan for user approval before running any qlog commands:

```
Goal: [title] · [near-term checkpoint] / [full arc]

Track 1: [Pillar Name]  (no dependencies)
  Milestone: [description] · due YYYY-MM-DD
  Resources:
    [paper] Title · ~45m · https://...
    [video] Title · ~60m · https://...
    [book]  Title · ~300m · https://...

Track 2: [Pillar Name]  (depends on: Track 1)
  Milestone: [description] · due YYYY-MM-DD
  Resources:
    [article] Title · ~20m · https://...
```

Ask the user to approve. They may:
- Approve in full → proceed to Phase 4
- Remove specific resources or tracks
- Adjust milestone descriptions or deadlines

**Do not run any qlog commands until the user explicitly approves.**

---

## Phase 4: Execution

Run qlog commands in the order below. Batch where possible using `;` in a single Bash call.

**Step 1: Check existing state**
```bash
qlog goal list && qlog track list
```
If the goal or tracks already exist, skip those creation steps.

**Step 2: Create the goal**
```bash
qlog goal new "Goal Title Here" --description "One-line description"
```
Note the slug printed in the output — you need it for subsequent steps.

**Step 3: Create tracks (leaves of dependency graph first)**
```bash
qlog track new "pillar-one" --description "Description" --goal <goal-slug> ; \
qlog track new "pillar-two" --description "Description" --goal <goal-slug>
```

**Step 4: Wire dependencies**
```bash
qlog track depends-on "dependent-track" "prerequisite-track"
```
Skip this step if there are no dependencies.

**Step 5: Add goal-level milestones**
```bash
qlog goal milestone add <goal-slug> --description "Milestone description" --deadline YYYY-MM-DD
```
Skip if no goal-level milestones were planned.

**Step 6: Add track milestones**
```bash
qlog track milestone add "pillar-one" --description "Milestone description" --deadline YYYY-MM-DD ; \
qlog track milestone add "pillar-two" --description "Milestone description" --deadline YYYY-MM-DD
```

**Step 7: Add resources (batch per track)**
```bash
qlog add --title "Paper Title" --type paper --track "pillar-one" --minutes 45 --url "https://..." ; \
qlog add --title "Video Title" --type video --track "pillar-one" --minutes 60 --url "https://..." ; \
qlog add --title "Book Title"  --type book  --track "pillar-one" --minutes 300 --url "https://..."
```
Repeat for each track.

**Step 8: Verify**
```bash
qlog goal show <goal-slug>
qlog track list
```
Show the output to the user.

---

## Notes for Future Development

**Output artifact tracking:** qlog currently tracks resources you *consume*. A future feature would add a category for things you *produce* — blog posts, side projects, conference talks — as public proof of expertise. The milestone `--artifact` flag can link to a resource ID today; the gap is a dedicated creation and tracking flow for output artifacts.
```

- [ ] **Step 2: Verify the file was written**

```bash
wc -l /Users/tom/.claude/skills/create-questlog/SKILL.md
```

Expected: ~160 lines.

- [ ] **Step 3: Verify all section headings are present**

```bash
grep "^## " /Users/tom/.claude/skills/create-questlog/SKILL.md
```

Expected output:
```
## Prerequisites
## qlog CLI Reference
## Phase 1: Goal Interview
## Phase 2: Curriculum Research
## Phase 3: Plan Presentation
## Phase 4: Execution
## Notes for Future Development
```

---

### Task 2: Verify spec coverage

**Files:**
- Read: `/Users/tom/Documents/personal/projects/questlog/docs/superpowers/specs/2026-04-23-questlog-goal-builder-skill-design.md`
- Read: `/Users/tom/.claude/skills/create-questlog/SKILL.md`

- [ ] **Step 1: Check spec requirements are met**

Read through both files and verify:

| Spec requirement | Where in SKILL.md |
|-----------------|-------------------|
| Goal interview extracts 4 fields (statement, horizon, baseline, domains) | Phase 1 table |
| Capability pillars → tracks | Phase 1 summary block |
| User confirms summary before research | Phase 1 final paragraph |
| Targeted web searches, specific not generic | Phase 2 search examples |
| 7 source packs: General Tech, AI/ML, Systems, Startup/Product, Web/Frontend, Courses & Video, Books | Phase 2 Source Packs |
| 5–10 resources per track, mixed types | Phase 2 Resource Quality Bar |
| Plan presented for approval before execution | Phase 3 |
| Execution order: goal → tracks → deps → milestones → resources → verify | Phase 4 steps 1–8 |
| Output artifact tracking noted for future | Notes for Future Development |

- [ ] **Step 2: Verify CLI commands in SKILL.md match actual qlog interface**

Cross-check the CLI reference block against the verified interface at the top of this plan:
- `qlog goal new "Title"` (positional arg, not flag) ✓
- `qlog track new <name> --goal <slug>` ✓
- `qlog track depends-on <track> <dep>` ✓
- `qlog track milestone add <track>` (not `goal milestone`) ✓
- `qlog goal milestone add <goal-slug>` ✓

If any command is wrong, correct it in SKILL.md and re-run Step 1 of this task.
