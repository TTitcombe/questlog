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
  1. [pillar-name]  (domain: ai-ml)
  2. [pillar-name]  (domain: systems)
  3. [pillar-name]  (domain: startup-product)
Domain packs activated: ai-ml, systems, startup-product
```

**Capability pillars** are 3–5 named areas that together constitute the goal. Each pillar becomes one track in qlog. The baseline shapes the curriculum — someone strong on fundamentals but weak on deployment gets different resources than someone starting from scratch.

Do not proceed to Phase 2 until the user confirms the summary.

---

## Phase 2: Curriculum Research

For each capability pillar, run targeted web searches to find 5–10 real, linkable resources. Use the source packs below based on the domain tags assigned to each pillar in Phase 1 — each domain tag name corresponds directly to a source pack below. Activate all packs whose tag appears in the Phase 1 summary. Multiple packs can be active simultaneously.

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

**If a search yields no useful results:** broaden the query (e.g. remove the year), try an alternative source from the same pack, or fall back to the General Tech pack. Do not hallucinate URLs — only include resources you can find with a real link.

---

## Phase 3: Plan Presentation

Present the full plan for user approval before running any qlog commands:

```
Goal: [title] · [near-term checkpoint] / [full arc]

Track 1: [Pillar Name]  (no dependencies)
  Milestone: [description] · due YYYY-MM-DD
  Resources:
    [paper] Title · ~45m · p1 · https://...
    [video] Title · ~60m · p2 · https://...
    [book]  Title · ~300m · p1 · https://...

Track 2: [Pillar Name]  (depends on: Track 1)
  Milestone: [description] · due YYYY-MM-DD
  Resources:
    [article] Title · ~20m · p[n] · https://...
```

(Resource list is abbreviated — include all 5–10 resources per track from Phase 2.)

Ask the user to approve. They may:
- Approve in full → proceed to Phase 4
- Remove specific resources or tracks
- Adjust milestone descriptions or deadlines

After incorporating any requested changes, re-present the updated plan and ask for final approval before proceeding to Phase 4.

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
The slug is the lowercased, hyphenated form of the title (e.g. "My AI Goal" → `my-ai-goal`). Verify it from the command output before using it in later steps.

**Step 3: Create tracks (leaves of dependency graph first)**
```bash
qlog track new "pillar-one" --description "Description" --goal <goal-slug> ; \
qlog track new "pillar-two" --description "Description" --goal <goal-slug>
```
Use lowercase kebab-case for track names (e.g. `"frontier-awareness"`, not `"Frontier Awareness"`). qlog treats the name as an identifier — spaces and capitals cause issues with subsequent commands.

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
qlog add --title "Paper Title" --type paper --track "pillar-one" --minutes 45 --url "https://..." --priority 1 ; \
qlog add --title "Video Title" --type video --track "pillar-one" --minutes 60 --url "https://..." --priority 2 ; \
qlog add --title "Book Title"  --type book  --track "pillar-one" --minutes 300 --url "https://..." --priority 1
```
Set priority 1–2 for foundational resources, 3–4 for intermediate, 5 for niche/advanced. Priority 0 means unset — avoid it as it breaks reading-order suggestions.
Repeat for each track.

**Step 8: Verify**
Run these separately so you can read each output clearly before showing the user.
```bash
qlog goal show <goal-slug>
qlog track list
```
Show the output to the user.

---

## Notes for Future Development (do not act on this section)

This is a developer note for future qlog enhancements — not an instruction for the current workflow.

**Output artifact tracking:** qlog currently tracks resources you *consume*. A future feature would add a category for things you *produce* — blog posts, side projects, conference talks — as public proof of expertise. The milestone `--artifact` flag can link to a resource ID today; the gap is a dedicated creation and tracking flow for output artifacts.
