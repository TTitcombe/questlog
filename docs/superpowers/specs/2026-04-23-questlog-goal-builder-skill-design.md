# questlog-goal-builder skill — Design Spec

**Date:** 2026-04-23
**Status:** Approved

---

## Overview

A Claude Code skill that conducts a guided conversation, researches real resources via web search, and executes `qlog` CLI commands to set up a fully structured goal with tracks, milestones, and a starter curriculum.

The skill is domain-agnostic for any technical career goal (AI/ML, systems, web, infra, startup-eng, etc.). It supersedes `create-questlog` for goal-structured use cases.

**Three phases:**
1. **Goal interview** — free conversation to establish goal, pillars, time horizon, and baseline
2. **Curriculum research** — targeted web search against domain-appropriate source packs; presents full plan for user approval
3. **Execution** — runs `qlog` CLI commands to build the structure; user can skip or modify at the batch level

---

## Phase 1: Goal Interview

The interview is conversational, not a form. The skill asks open questions, listens, and synthesises.

**Four things to extract:**

| Field | Description |
|-------|-------------|
| Goal statement | What the user wants to be able to do / be known for |
| Time horizon | Near-term milestone (3–6 months) and full arc (1–2 years) |
| Current baseline | What's strong, what's rusty, what's missing — shapes curriculum difficulty |
| Domain(s) | Inferred from conversation; used to select source packs; can be multiple |

After the conversation, the skill presents a structured summary — goal title, capability pillars, domain tags — and asks for confirmation before moving to research.

**Capability pillars** are the key output: 3–5 named areas that together constitute the goal. These map directly to tracks in qlog. The skill suggests them; the user approves or edits.

The baseline matters because it shapes the curriculum — someone strong on ML fundamentals but weak on deployment gets different resources than someone starting from scratch.

---

## Phase 2: Curriculum Research

For each capability pillar, the skill runs targeted web searches against the appropriate source packs and finds 5–10 concrete, linkable resources. Searches are specific ("best course for production LLM deployment 2025") not generic ("AI resources").

Resources get typed and time-estimated to land correctly in qlog (`--type paper --minutes 45`, etc.).

### Source Packs

Packs are activated by domain detection and are stackable (a goal can activate multiple packs).

| Pack | Key sources |
|------|-------------|
| General tech | HackerNews, engineering blogs (Stripe, Cloudflare, Netflix, Uber), MIT/Stanford OCW, official docs |
| AI/ML | Hugging Face Daily Papers, Papers With Code, fast.ai, DeepLearning.ai, lab blogs (Anthropic, OpenAI, DeepMind, Meta AI) |
| Systems/Infra | USENIX, High Scalability, Brendan Gregg, ACM |
| Startup/Product | First Round Review, Lenny's Newsletter, Paul Graham essays |
| Web/Frontend | MDN, framework official blogs, web.dev |
| Courses & Video | YouTube (domain-specific channels, e.g. Karpathy, Yannic Kilcher, Fireship), Coursera, edX, conference recordings |
| Books | O'Reilly (primary for tech), Manning, Pragmatic Programmers, MIT Press |

Domain-specific channels are also added to their relevant pack (e.g. Yannic Kilcher / 3Blue1Brown in AI/ML; CppCon recordings in Systems).

### Plan Presentation

After research, the skill presents the full plan before touching qlog:

```
Goal: [title] · [timeline]

Track: [Pillar Name]  (dependencies: none)
  Milestone: [description] · due [date]
  Resources:
    [paper] Title · ~45m · url
    [video] Title · ~60m · url
    ...

Track: [Pillar Name]  (depends on: [track])
  Milestone: [description] · due [date]
  Resources: ...
```

The user approves the whole structure, then can flag individual items to drop before execution begins.

---

## Phase 3: Execution

Commands run in dependency order (no track is created before its dependencies):

1. `qlog goal new` — creates the goal
2. `qlog track new --goal <slug>` — one per track, leaves of the dependency graph first
3. `qlog track depends-on` — wires up inter-track dependencies
4. `qlog track milestone add --deadline <date>` — milestones per track
5. `qlog add --track <name> --type <type> --minutes <n> --url <url>` — seed resources

**User control:** The skill batches by stage and asks for confirmation at each ("Create these 4 tracks?") rather than prompting per-command. The user can skip an entire stage or flag individual items to drop.

**Idempotency:** If a goal or track already exists, the skill detects the error and skips rather than failing — safe to re-run.

**Resource titles:** The skill uses clean, descriptive titles (not URL-derived slugs) so qlog file names remain human-readable.

---

## Future Work

### Output Artifact Tracking

qlog currently models resources as things you *consume*. A complete career goal also requires tracking things you *produce* — blog posts, side projects, conference talks, OSS contributions — as public proof of expertise.

The milestone system already supports `--artifact <resource-id>` for linking proof; the gap is a creation and tracking flow for output artifacts. This is a separate qlog feature to design and implement.

Suggested future resource types or categories:
- `post` — blog post or LinkedIn article written
- `project` — side project or OSS contribution
- `talk` — meetup or conference presentation

---

## Skill Location

Lives in the superpowers plugins directory alongside `create-questlog`. For goal-structured use cases (which is the primary use case), this skill should be used in preference to `create-questlog`.
