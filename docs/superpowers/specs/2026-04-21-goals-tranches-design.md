# Goals & Tracks Design

**Date:** 2026-04-21  
**Status:** Approved  
**Scope:** Data model + CLI for goal-setting with linked competency tracks

---

## Terminology

- **Track** — the user-facing term throughout all CLI commands and output. Tracks are the competency areas that make up a goal.
- **Tranche** — used only in this spec as a conceptual shorthand when distinguishing a goal-linked track from a standalone track. Never appears in the CLI or user-facing output.

---

## Problem

questlog is good at collecting and working through resources, but has no concept of a *destination*. There is no way to define a transformational learning goal (e.g. "become an AI research engineer"), break it into competency areas, sequence them, set milestones, and measure progress toward the overall target. This design adds that layer without disrupting the existing standalone-track use case.

---

## Hierarchy

```
Goal  (directional, optional deadline via milestones)
  ├── Goal Milestones  ("targets" — measurable, time-boxed checkpoints)
  └── Tracks  (optionally linked to a goal)
        ├── Core resources   (non-negotiable, gate track completion)
        ├── Extra resources  (enrichment, no gate effect)
        └── Track Milestones  (completion criteria: self-reports or linked artifacts)
```

Tracks that are not linked to a goal continue to work exactly as today.

---

## Data Model

### Goal (`~/.questlog/goals/<slug>/goal.json`)

```json
{
  "slug": "ai-research-engineer",
  "title": "AI Research Engineer at a frontier lab",
  "description": "Transition from SWE to AI research at a frontier lab.",
  "created": "2026-04-21T00:00:00Z",
  "milestones": [
    {
      "id": "interview-ready",
      "description": "Pass a research internship interview",
      "deadline": "2026-09-01",
      "artifact_resource_id": null,
      "completed_at": null
    }
  ]
}
```

Fields:
- `slug` — derived from title, used as directory name and foreign key
- `description` — optional, set via `--description` flag on creation
- `milestones` — goal-level checkpoints ("targets"); each has an optional deadline and optional link to a proof artifact resource

### Track (`~/.questlog/tracks/<name>/track.json`) — extended

New optional fields (all omitted for standalone tracks):

```json
{
  "name": "transformer-theory",
  "goal_slug": "ai-research-engineer",
  "depends_on": ["linear-algebra", "probability"],
  "milestones": [
    {
      "id": "can-implement-attention",
      "description": "Can implement multi-head attention from scratch",
      "deadline": "2026-07-01",
      "artifact_resource_id": "my-attention-implementation",
      "completed_at": null
    }
  ]
}
```

Fields:
- `goal_slug` — links track to a goal; omitted for standalone tracks
- `depends_on` — list of track slugs that must be complete before this track is available; `[]` means available immediately (parallel with other tracks)
- `milestones` — track-level completion criteria

The `depends_on` field forms a DAG. Cycle detection is enforced on write. A track with `depends_on: []` is available in parallel with any other track in the goal.

### Resource — extended

One new field:

```go
IsCore bool `yaml:"is_core,omitempty"`
```

Core resources are non-negotiable for track completion. Extra resources (default) are enrichment only. Set via `--core` flag on `qlog add` or `qlog classify`.

### Milestone (shared struct, used at both goal and track level)

```go
type Milestone struct {
    ID               string     `json:"id"`
    Description      string     `json:"description"`
    Deadline         *time.Time `json:"deadline,omitempty"`
    ArtifactResource string     `json:"artifact_resource_id,omitempty"`
    CompletedAt      *time.Time `json:"completed_at,omitempty"`
}
```

`ArtifactResource` optionally links to a questlog resource the user created as proof (a note, project writeup, derivation, etc.). A milestone is complete when `CompletedAt` is set.

---

## Progress Computation

All progress is computed on read from existing data — no stored rollup fields.

**Track progress** uses two independent signals:
1. **Core resource completion** — `done` core resources ÷ total core resources (0–100%). A track with no core resources flagged shows a warning rather than reporting 100%.
2. **Milestone completion** — count of milestones with `completed_at` set.

**Track complete** when: all core resources are `status: done` AND all milestones have `completed_at` set AND at least one of (core resources, milestones) is non-empty. A track with nothing configured is never considered complete — it won't unblock dependents until the user adds and completes at least one signal.

**Track availability** (derived from DAG):
- `locked` — one or more `depends_on` tracks are not yet complete
- `available` — all dependencies complete, or no dependencies

**Goal progress** = percentage of tracks complete. Goal-level milestones are tracked independently and do not affect the percentage.

---

## CLI Commands

### Goal management

```
qlog goal new "<title>" [--description "<desc>"]
qlog goal list
qlog goal show <slug>
qlog goal milestone add <slug>        # interactive: description, optional deadline
qlog goal milestone done <slug> <id>  # marks complete, optionally links artifact
```

### Linking tracks to goals

```
qlog track new <name> --goal <slug>        # create track already linked to a goal
qlog track set-goal <name> <goal-slug>     # link an existing track to a goal
qlog track depends-on <name> <other>       # add a dependency edge (validates no cycles)
qlog track milestone add <name>            # interactive: description, optional deadline
qlog track milestone done <name> <id>      # marks complete, optionally links artifact resource
```

### Core resource flag

```
qlog add <url> --track <name> --core       # mark as core on add
qlog classify <id>                         # existing command, gains --core flag
```

### Progress view

```
qlog progress <goal-slug>
```

Example output:

```
AI Research Engineer  ·  interview-ready by Sep 2026  [47% core done]

  ✓ linear-algebra      3/3 core  · 1/1 milestones ✓
  ◑ probability         2/4 core  · deadline Jun 1
  ○ transformer-theory  locked (needs: linear-algebra, probability)
  ◑ gpu-programming     1/2 core
  ○ reading-papers      0/3 core
  ○ c-basics            0/2 core
```

Locked tracks display their unmet dependencies. Available tracks show core resource and milestone progress.

### Focus — updated behaviour

`qlog focus` gains goal-awareness when run without `--track`. If the user has any active goals, focus surfaces goal context at the top of the session, then filters resources to available tracks only (tracks whose dependencies are all complete).

Within available tracks, resources are ordered: in-progress first, then unread. Done resources are excluded as today.

Pending milestones — both goal-level and track-level — are highlighted at the top of the focus session to remind the user of upcoming checkpoints.

Example focus header (new):

```
Goal: AI Research Engineer  [47% core done]  · interview-ready by Sep 2026 ⚑
Available tracks: probability, gpu-programming, reading-papers

! Track milestone due Jun 1: "Can explain gradient flow through attention"

─────────────────────────────────────────────────────
```

If `--track` is specified, focus behaves exactly as today (no goal context injected). If multiple goals exist, focus picks the goal with the nearest milestone deadline; a future `--goal` flag can make this explicit.

---

## What is not in scope

- Resource discovery / AI-assisted curation (future)
- Self-evaluation / quizzes (future)
- Spaced repetition / retention (future)
- TUI or Mac app (future)
- Community-shared goal templates (future)
- `--goal` flag on `qlog focus` (future, once multi-goal use is common)
