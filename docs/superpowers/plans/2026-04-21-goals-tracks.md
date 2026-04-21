# Goals & Tracks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Goals as a first-class entity that links existing tracks into sequenced learning plans, with core-resource gating, milestones, and progress computation.

**Architecture:** Goals live at `~/.questlog/goals/<slug>/goal.json`. Tracks gain optional `goal_slug`, `depends_on`, and `milestones` fields — all zero-value for standalone tracks, so nothing breaks. Progress is computed on-the-fly from resource statuses and milestone `completed_at` fields; no stored rollups.

**Tech Stack:** Go stdlib, cobra, promptui, charmbracelet/bubbletea (existing deps). No new dependencies.

**Spec:** `docs/superpowers/specs/2026-04-21-goals-tranches-design.md`

---

## File Map

| File | Action | Purpose |
|------|--------|---------|
| `internal/model/goal.go` | Create | Goal + Milestone types |
| `internal/model/track.go` | Modify | Add GoalSlug, DependsOn, Milestones fields |
| `internal/model/resource.go` | Modify | Add IsCore field |
| `internal/model/progress.go` | Create | Pure progress computation + DAG cycle detection |
| `internal/model/progress_test.go` | Create | Tests for progress computation and HasCycle |
| `internal/store/slug.go` | Modify | Export Slugify for CLI use |
| `internal/store/goal.go` | Create | SaveGoal, LoadGoal, ListGoals on FSStore |
| `internal/store/store.go` | Modify | Add Goal methods + SaveTrack to Store interface |
| `internal/store/fs.go` | Modify | Implement SaveTrack, init goals dir in New() |
| `internal/store/markdown.go` | Modify | Add is_core to frontmatter struct |
| `internal/cli/goal.go` | Create | qlog goal new/list/show/milestone add/done |
| `internal/cli/progress.go` | Create | qlog progress |
| `internal/cli/track.go` | Modify | Add set-goal, depends-on, milestone add/done; --goal on track new |
| `internal/cli/add.go` | Modify | Add --core flag |
| `internal/cli/classify.go` | Modify | Add --core flag |
| `internal/cli/focus.go` | Modify | Goal-awareness when --track not specified |
| `internal/cli/focus_tui.go` | Modify | Render goal header + milestone alerts |
| `internal/cli/root.go` | Modify | Register goal and progress commands |

---

## Task 1: Model types

**Files:**
- Create: `internal/model/goal.go`
- Modify: `internal/model/track.go`
- Modify: `internal/model/resource.go`

- [ ] **Step 1: Create `internal/model/goal.go`**

```go
package model

import "time"

type Goal struct {
	Slug        string      `json:"slug"`
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Created     time.Time   `json:"created"`
	Milestones  []Milestone `json:"milestones,omitempty"`
}

type Milestone struct {
	ID               string     `json:"id"`
	Description      string     `json:"description"`
	Deadline         *time.Time `json:"deadline,omitempty"`
	ArtifactResource string     `json:"artifact_resource_id,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}
```

- [ ] **Step 2: Extend `internal/model/track.go`**

Replace the existing struct with:

```go
package model

import "time"

type Track struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Created     time.Time   `json:"created"`
	GoalSlug    string      `json:"goal_slug,omitempty"`
	DependsOn   []string    `json:"depends_on,omitempty"`
	Milestones  []Milestone `json:"milestones,omitempty"`
}
```

- [ ] **Step 3: Add IsCore to `internal/model/resource.go`**

Add `IsCore` after the `Priority` field in the `Resource` struct:

```go
IsCore   bool         // marks resource as non-negotiable for track completion
```

- [ ] **Step 4: Build and verify no compile errors**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 2: Export Slugify

**Files:**
- Modify: `internal/store/slug.go`

The `slugify` function is unexported. CLI code needs it for milestone IDs.

- [ ] **Step 1: Add exported wrapper to `internal/store/slug.go`**

Add after the existing `uniqueSlug` function:

```go
// Slugify is the exported variant of slugify for use outside this package.
func Slugify(title string) string {
	return slugify(title)
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 3: Store — Goal CRUD, SaveTrack, interface updates

**Files:**
- Modify: `internal/store/store.go`
- Create: `internal/store/goal.go`
- Modify: `internal/store/fs.go`

- [ ] **Step 1: Add Goal methods and SaveTrack to `internal/store/store.go`**

Add to the `Store` interface after the existing Track operations block:

```go
// Goal operations
SaveGoal(g model.Goal) error
LoadGoal(slug string) (model.Goal, error)
ListGoals() ([]model.Goal, error)

// SaveTrack updates an existing track's metadata (goal link, deps, milestones).
SaveTrack(t model.Track) error
```

- [ ] **Step 2: Create `internal/store/goal.go`**

```go
package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/TTitcombe/questlog/internal/model"
)

func (s *FSStore) goalDir(slug string) string {
	return filepath.Join(s.dataDir, "goals", slug)
}

func (s *FSStore) goalPath(slug string) string {
	return filepath.Join(s.goalDir(slug), "goal.json")
}

func (s *FSStore) SaveGoal(g model.Goal) error {
	if err := os.MkdirAll(s.goalDir(g.Slug), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.goalPath(g.Slug), data, 0644)
}

func (s *FSStore) LoadGoal(slug string) (model.Goal, error) {
	data, err := os.ReadFile(s.goalPath(slug))
	if err != nil {
		return model.Goal{}, err
	}
	var g model.Goal
	return g, json.Unmarshal(data, &g)
}

func (s *FSStore) ListGoals() ([]model.Goal, error) {
	dir := filepath.Join(s.dataDir, "goals")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var goals []model.Goal
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		g, err := s.LoadGoal(e.Name())
		if err != nil {
			continue
		}
		goals = append(goals, g)
	}
	return goals, nil
}
```

- [ ] **Step 3: Add SaveTrack and goals dir init to `internal/store/fs.go`**

In `New()`, add `filepath.Join(dataDir, "goals")` to the `dirs` slice:

```go
dirs := []string{
    dataDir,
    filepath.Join(dataDir, "tracks"),
    filepath.Join(dataDir, "inbox"),
    filepath.Join(dataDir, "goals"),
}
```

Add `SaveTrack` after `ListTracks`:

```go
func (s *FSStore) SaveTrack(t model.Track) error {
	path := filepath.Join(s.dataDir, "tracks", t.Name, "track.json")
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 4: Frontmatter — is_core field

**Files:**
- Modify: `internal/store/markdown.go`

- [ ] **Step 1: Add IsCore to the `frontmatter` struct**

Add after the `Priority` field:

```go
IsCore bool `yaml:"is_core,omitempty"`
```

- [ ] **Step 2: Populate IsCore in `parseMarkdown`**

In the `return model.Resource{...}` block, add:

```go
IsCore:           fm.IsCore,
```

- [ ] **Step 3: Populate IsCore in `marshalMarkdown`**

In the `fm := frontmatter{...}` block, add:

```go
IsCore: r.IsCore,
```

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 5: Progress computation + tests

**Files:**
- Create: `internal/model/progress.go`
- Create: `internal/model/progress_test.go`

- [ ] **Step 1: Write the failing tests — `internal/model/progress_test.go`**

```go
package model_test

import (
	"testing"
	"time"

	"github.com/TTitcombe/questlog/internal/model"
)

func timePtr(t time.Time) *time.Time { return &t }

func TestHasCycle_NoCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{}},
		{Name: "b", DependsOn: []string{"a"}},
		{Name: "c", DependsOn: []string{"a"}},
	}
	if model.HasCycle(tracks) {
		t.Fatal("expected no cycle")
	}
}

func TestHasCycle_WithCycle(t *testing.T) {
	tracks := []model.Track{
		{Name: "a", DependsOn: []string{"b"}},
		{Name: "b", DependsOn: []string{"a"}},
	}
	if !model.HasCycle(tracks) {
		t.Fatal("expected cycle to be detected")
	}
}

func TestComputeGoalProgress_ParallelTracks(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	done := timePtr(time.Now())
	tracks := []model.Track{
		{
			Name: "maths", GoalSlug: "g", DependsOn: []string{},
			Milestones: []model.Milestone{{ID: "m1", Description: "done", CompletedAt: done}},
		},
		{Name: "coding", GoalSlug: "g", DependsOn: []string{}},
	}
	resources := map[string][]model.Resource{
		"maths":  {{IsCore: true, Status: model.StatusDone}},
		"coding": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	if gp.TracksComplete != 1 {
		t.Errorf("expected 1 track complete, got %d", gp.TracksComplete)
	}
	if gp.TracksTotal != 2 {
		t.Errorf("expected 2 tracks total, got %d", gp.TracksTotal)
	}

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "maths" && tp.Status != model.TrackStatusComplete {
			t.Errorf("maths should be complete")
		}
		if tp.Track.Name == "coding" && tp.Status != model.TrackStatusAvailable {
			t.Errorf("coding should be available, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_LockedByDep(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	tracks := []model.Track{
		{Name: "a", GoalSlug: "g", DependsOn: []string{}},
		{Name: "b", GoalSlug: "g", DependsOn: []string{"a"}},
	}
	resources := map[string][]model.Resource{
		"a": {{IsCore: true, Status: model.StatusUnread}},
		"b": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "b" && tp.Status != model.TrackStatusLocked {
			t.Errorf("b should be locked while a is incomplete, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_UnlockedAfterDepComplete(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	done := timePtr(time.Now())
	tracks := []model.Track{
		{
			Name: "a", GoalSlug: "g", DependsOn: []string{},
			Milestones: []model.Milestone{{ID: "m1", Description: "x", CompletedAt: done}},
		},
		{Name: "b", GoalSlug: "g", DependsOn: []string{"a"}},
	}
	resources := map[string][]model.Resource{
		"a": {{IsCore: true, Status: model.StatusDone}},
		"b": {{IsCore: true, Status: model.StatusUnread}},
	}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	for _, tp := range gp.Tracks {
		if tp.Track.Name == "b" && tp.Status != model.TrackStatusAvailable {
			t.Errorf("b should be available once a is complete, got %v", tp.Status)
		}
	}
}

func TestComputeGoalProgress_NoSignalsNeverComplete(t *testing.T) {
	goal := model.Goal{Slug: "g", Title: "G"}
	tracks := []model.Track{
		{Name: "empty", GoalSlug: "g", DependsOn: []string{}},
	}
	resources := map[string][]model.Resource{"empty": {}}

	gp := model.ComputeGoalProgress(goal, tracks, resources)

	if gp.TracksComplete != 0 {
		t.Error("track with no signals should never be complete")
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/model/...
```

Expected: compile error — `model.HasCycle`, `model.ComputeGoalProgress`, `model.TrackStatus*` undefined.

- [ ] **Step 3: Create `internal/model/progress.go`**

```go
package model

// TrackStatus describes a track's availability within a goal.
type TrackStatus int

const (
	TrackStatusLocked    TrackStatus = iota // has incomplete dependencies
	TrackStatusAvailable                    // deps met, not yet complete
	TrackStatusComplete                     // all signals satisfied
)

// TrackProgress is the computed state for one goal track.
type TrackProgress struct {
	Track          Track
	Status         TrackStatus
	LockedBy       []string // names of incomplete dependency tracks
	CoreDone       int
	CoreTotal      int
	MilestoneDone  int
	MilestoneTotal int
}

// GoalProgress is the computed state for a full goal.
type GoalProgress struct {
	Goal           Goal
	TracksComplete int
	TracksTotal    int
	Tracks         []TrackProgress
}

// ComputeGoalProgress derives progress state from raw data. resources maps
// track name → resources in that track.
func ComputeGoalProgress(goal Goal, tracks []Track, resources map[string][]Resource) GoalProgress {
	// Iteratively mark tracks complete when deps are done and signals are satisfied.
	complete := make(map[string]bool, len(tracks))
	for {
		changed := false
		for _, t := range tracks {
			if complete[t.Name] {
				continue
			}
			if trackInternallyComplete(t, resources[t.Name]) && depsAllComplete(t, complete) {
				complete[t.Name] = true
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	tps := make([]TrackProgress, 0, len(tracks))
	for _, t := range tracks {
		tps = append(tps, buildTrackProgress(t, resources[t.Name], complete))
	}

	done := 0
	for _, tp := range tps {
		if tp.Status == TrackStatusComplete {
			done++
		}
	}

	return GoalProgress{
		Goal:           goal,
		TracksComplete: done,
		TracksTotal:    len(tracks),
		Tracks:         tps,
	}
}

func trackInternallyComplete(t Track, resources []Resource) bool {
	coreTotal := 0
	for _, r := range resources {
		if r.IsCore {
			coreTotal++
		}
	}
	// A track with no signals configured can never be complete.
	if coreTotal == 0 && len(t.Milestones) == 0 {
		return false
	}
	for _, r := range resources {
		if r.IsCore && r.Status != StatusDone {
			return false
		}
	}
	for _, m := range t.Milestones {
		if m.CompletedAt == nil {
			return false
		}
	}
	return true
}

func depsAllComplete(t Track, complete map[string]bool) bool {
	for _, dep := range t.DependsOn {
		if !complete[dep] {
			return false
		}
	}
	return true
}

func buildTrackProgress(t Track, resources []Resource, complete map[string]bool) TrackProgress {
	coreDone, coreTotal := 0, 0
	for _, r := range resources {
		if r.IsCore {
			coreTotal++
			if r.Status == StatusDone {
				coreDone++
			}
		}
	}

	milestoneDone := 0
	for _, m := range t.Milestones {
		if m.CompletedAt != nil {
			milestoneDone++
		}
	}

	var status TrackStatus
	var lockedBy []string

	if complete[t.Name] {
		status = TrackStatusComplete
	} else {
		for _, dep := range t.DependsOn {
			if !complete[dep] {
				lockedBy = append(lockedBy, dep)
			}
		}
		if len(lockedBy) > 0 {
			status = TrackStatusLocked
		} else {
			status = TrackStatusAvailable
		}
	}

	return TrackProgress{
		Track:          t,
		Status:         status,
		LockedBy:       lockedBy,
		CoreDone:       coreDone,
		CoreTotal:      coreTotal,
		MilestoneDone:  milestoneDone,
		MilestoneTotal: len(t.Milestones),
	}
}

// HasCycle returns true if the depends_on graph among tracks contains a cycle.
// Uses Kahn's algorithm (topological sort).
func HasCycle(tracks []Track) bool {
	inDegree := make(map[string]int, len(tracks))
	adj := make(map[string][]string, len(tracks))

	for _, t := range tracks {
		if _, ok := inDegree[t.Name]; !ok {
			inDegree[t.Name] = 0
		}
		for _, dep := range t.DependsOn {
			inDegree[t.Name]++
			adj[dep] = append(adj[dep], t.Name)
		}
	}

	queue := make([]string, 0, len(tracks))
	for _, t := range tracks {
		if inDegree[t.Name] == 0 {
			queue = append(queue, t.Name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adj[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	return visited < len(tracks)
}
```

- [ ] **Step 4: Run tests and verify they pass**

```bash
go test ./internal/model/... -v
```

Expected: all 6 tests PASS.

---

## Task 6: CLI — goal commands

**Files:**
- Create: `internal/cli/goal.go`

- [ ] **Step 1: Create `internal/cli/goal.go`**

```go
package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newGoalCmd(getStore func() *store.FSStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goal",
		Short: "Manage learning goals",
	}
	cmd.AddCommand(
		newGoalNewCmd(getStore),
		newGoalListCmd(getStore),
		newGoalShowCmd(getStore),
		newGoalMilestoneCmd(getStore),
	)
	return cmd
}

func newGoalNewCmd(getStore func() *store.FSStore) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new learning goal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			title := args[0]
			slug := store.Slugify(title)

			if _, err := s.LoadGoal(slug); err == nil {
				return fmt.Errorf("goal %q already exists", slug)
			}

			g := model.Goal{
				Slug:        slug,
				Title:       title,
				Description: description,
				Created:     time.Now(),
			}
			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Created goal: %s\n", ui.Success.Render("✓"), ui.Highlight.Render(title))
			fmt.Printf("  Slug: %s\n", ui.Dim.Render(slug))
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "goal description")
	return cmd
}

func newGoalListCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all goals",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			goals, err := s.ListGoals()
			if err != nil {
				return err
			}
			if len(goals) == 0 {
				fmt.Println(ui.Muted.Render("No goals yet. Create one with: qlog goal new \"<title>\""))
				return nil
			}
			tracks, _ := s.ListTracks()
			for _, g := range goals {
				linked := 0
				for _, t := range tracks {
					if t.GoalSlug == g.Slug {
						linked++
					}
				}
				fmt.Printf("  %s  %s\n", ui.Highlight.Render(g.Title), ui.Dim.Render(g.Slug))
				if g.Description != "" {
					fmt.Printf("    %s\n", ui.Muted.Render(g.Description))
				}
				pending := 0
				for _, m := range g.Milestones {
					if m.CompletedAt == nil {
						pending++
					}
				}
				fmt.Printf("    %d tracks · %d milestones (%d pending)\n\n",
					linked, len(g.Milestones), pending)
			}
			return nil
		},
	}
}

func newGoalShowCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "show <slug>",
		Short: "Show goal details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
			}

			fmt.Printf("%s\n", ui.Bold.Render(g.Title))
			if g.Description != "" {
				fmt.Printf("%s\n", ui.Muted.Render(g.Description))
			}
			fmt.Printf("Slug: %s  Created: %s\n\n", ui.Dim.Render(g.Slug), g.Created.Format("2006-01-02"))

			if len(g.Milestones) > 0 {
				fmt.Println(ui.Bold.Render("Goal milestones:"))
				for _, m := range g.Milestones {
					check := "○"
					if m.CompletedAt != nil {
						check = ui.Success.Render("✓")
					}
					deadline := ""
					if m.Deadline != nil {
						deadline = ui.Warning.Render(fmt.Sprintf("  due %s", m.Deadline.Format("2006-01-02")))
					}
					fmt.Printf("  %s %s%s\n", check, m.Description, deadline)
					fmt.Printf("    %s\n", ui.Dim.Render(m.ID))
				}
				fmt.Println()
			}

			tracks, _ := s.ListTracks()
			fmt.Println(ui.Bold.Render("Tracks:"))
			found := false
			for _, t := range tracks {
				if t.GoalSlug != g.Slug {
					continue
				}
				found = true
				deps := ""
				if len(t.DependsOn) > 0 {
					deps = ui.Dim.Render(fmt.Sprintf("  needs: %s", strings.Join(t.DependsOn, ", ")))
				}
				fmt.Printf("  %s%s\n", ui.Highlight.Render(t.Name), deps)
			}
			if !found {
				fmt.Println(ui.Muted.Render("  No tracks linked yet. Use: qlog track set-goal <name> " + g.Slug))
			}
			return nil
		},
	}
}

func newGoalMilestoneCmd(getStore func() *store.FSStore) *cobra.Command {
	ms := &cobra.Command{
		Use:   "milestone",
		Short: "Manage goal milestones",
	}
	ms.AddCommand(
		newGoalMilestoneAddCmd(getStore),
		newGoalMilestoneDoneCmd(getStore),
	)
	return ms
}

func newGoalMilestoneAddCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		description string
		deadline    string
		artifact    string
	)
	cmd := &cobra.Command{
		Use:   "add <goal-slug>",
		Short: "Add a milestone to a goal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
			}

			if description == "" {
				description = mustPrompt("Milestone description", "", func(v string) error {
					if strings.TrimSpace(v) == "" {
						return fmt.Errorf("description cannot be empty")
					}
					return nil
				})
			}

			m := model.Milestone{
				ID:               store.Slugify(description),
				Description:      description,
				ArtifactResource: artifact,
			}
			if deadline != "" {
				d, err := time.Parse("2006-01-02", deadline)
				if err != nil {
					return fmt.Errorf("deadline must be YYYY-MM-DD, got %q", deadline)
				}
				m.Deadline = &d
			}

			g.Milestones = append(g.Milestones, m)
			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Added milestone: %s\n", ui.Success.Render("✓"), m.Description)
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "milestone description")
	cmd.Flags().StringVar(&deadline, "deadline", "", "target date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID that proves this milestone")
	return cmd
}

func newGoalMilestoneDoneCmd(getStore func() *store.FSStore) *cobra.Command {
	var artifact string
	cmd := &cobra.Command{
		Use:   "done <goal-slug> <milestone-id>",
		Short: "Mark a goal milestone complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
			}

			found := false
			now := time.Now()
			for i, m := range g.Milestones {
				if m.ID == args[1] {
					g.Milestones[i].CompletedAt = &now
					if artifact != "" {
						g.Milestones[i].ArtifactResource = artifact
					}
					found = true
					break
				}
			}
			if !found {
				// Print available IDs
				ids := make([]string, len(g.Milestones))
				for i, m := range g.Milestones {
					ids[i] = m.ID
				}
				data, _ := json.Marshal(ids)
				return fmt.Errorf("milestone %q not found; available: %s", args[1], data)
			}

			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Milestone complete: %s\n", ui.Success.Render("✓"), args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 7: CLI — progress command

**Files:**
- Create: `internal/cli/progress.go`

- [ ] **Step 1: Create `internal/cli/progress.go`**

```go
package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newProgressCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "progress <goal-slug>",
		Short: "Show progress toward a goal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			slug := args[0]

			g, err := s.LoadGoal(slug)
			if err != nil {
				return fmt.Errorf("goal %q not found", slug)
			}

			allTracks, err := s.ListTracks()
			if err != nil {
				return err
			}

			var goalTracks []model.Track
			for _, t := range allTracks {
				if t.GoalSlug == slug {
					goalTracks = append(goalTracks, t)
				}
			}

			resources := make(map[string][]model.Resource, len(goalTracks))
			for _, t := range goalTracks {
				res, err := s.ListResources(store.ResourceFilter{Track: t.Name})
				if err != nil {
					return err
				}
				resources[t.Name] = res
			}

			gp := model.ComputeGoalProgress(g, goalTracks, resources)

			// Header
			pendingMilestone := ""
			for _, m := range g.Milestones {
				if m.CompletedAt == nil && m.Deadline != nil {
					pendingMilestone = fmt.Sprintf("  · %s by %s", m.Description, m.Deadline.Format("Jan 2006"))
					break
				}
			}
			pct := 0
			if gp.TracksTotal > 0 {
				pct = gp.TracksComplete * 100 / gp.TracksTotal
			}
			fmt.Printf("%s%s  %s\n\n",
				ui.Bold.Render(g.Title),
				pendingMilestone,
				ui.Dim.Render(fmt.Sprintf("[%d/%d tracks complete · %d%%]", gp.TracksComplete, gp.TracksTotal, pct)),
			)

			// Tracks
			for _, tp := range gp.Tracks {
				icon := "○"
				switch tp.Status {
				case model.TrackStatusComplete:
					icon = ui.Success.Render("✓")
				case model.TrackStatusAvailable:
					icon = "◑"
				case model.TrackStatusLocked:
					icon = ui.Muted.Render("○")
				}

				core := ""
				if tp.CoreTotal > 0 {
					core = fmt.Sprintf("  %d/%d core", tp.CoreDone, tp.CoreTotal)
				} else {
					core = ui.Warning.Render("  no core resources set")
				}

				milestones := ""
				if tp.MilestoneTotal > 0 {
					milestones = fmt.Sprintf("  · %d/%d milestones", tp.MilestoneDone, tp.MilestoneTotal)
				}

				locked := ""
				if tp.Status == model.TrackStatusLocked {
					locked = ui.Muted.Render(fmt.Sprintf("  (needs: %s)", strings.Join(tp.LockedBy, ", ")))
				}

				// Upcoming deadline from track milestones
				deadline := ""
				for _, m := range tp.Track.Milestones {
					if m.CompletedAt == nil && m.Deadline != nil {
						deadline = ui.Warning.Render(fmt.Sprintf("  ⚑ due %s", m.Deadline.Format("Jan 2 2006")))
						break
					}
				}

				fmt.Printf("  %s  %-24s%s%s%s%s\n",
					icon,
					ui.Highlight.Render(tp.Track.Name),
					core,
					milestones,
					locked,
					deadline,
				)
			}

			if len(goalTracks) == 0 {
				fmt.Println(ui.Muted.Render("  No tracks linked. Use: qlog track set-goal <name> " + slug))
			}

			fmt.Println()
			return nil
		},
	}
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 8: Register new commands in root.go

**Files:**
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Register goal and progress commands**

Open `internal/cli/root.go`. Find the `root.AddCommand(...)` block and add:

```go
root.AddCommand(newGoalCmd(getStore))
root.AddCommand(newProgressCmd(getStore))
```

alongside the existing command registrations.

- [ ] **Step 2: Build and smoke-test**

```bash
go build ./... && go run ./cmd/qlog/main.go --help
```

Expected: `goal` and `progress` appear in the command list.

```bash
go run ./cmd/qlog/main.go goal --help
```

Expected: shows `new`, `list`, `show`, `milestone` subcommands.

---

## Task 9: Extend track commands

**Files:**
- Modify: `internal/cli/track.go`

- [ ] **Step 1: Add `--goal` flag to `newTrackNewCmd`**

In `newTrackNewCmd`, add a `goalSlug` variable and flag:

```go
var goalSlug string
```

```go
cmd.Flags().StringVar(&goalSlug, "goal", "", "link this track to a goal (goal slug)")
```

In the `RunE`, after building the `model.Track`, set `GoalSlug` if provided:

```go
t := model.Track{
    Name:        name,
    Description: description,
    Tags:        tagList,
    Created:     time.Now(),
    GoalSlug:    goalSlug,
}
```

If `goalSlug` is set, verify the goal exists before creating:

```go
if goalSlug != "" {
    if _, err := s.LoadGoal(goalSlug); err != nil {
        return fmt.Errorf("goal %q not found; create it first with: qlog goal new", goalSlug)
    }
}
```

- [ ] **Step 2: Add `set-goal`, `depends-on`, `milestone add`, `milestone done` to `newTrackCmd`**

In `newTrackCmd`, extend the `AddCommand` list:

```go
track.AddCommand(
    newTrackNewCmd(getStore),
    newTrackListCmd(getStore),
    newTrackShowCmd(getStore),
    newTrackSetGoalCmd(getStore),
    newTrackDependsOnCmd(getStore),
    newTrackMilestoneCmd(getStore),
)
```

- [ ] **Step 3: Add the new subcommand functions to `internal/cli/track.go`**

```go
func newTrackSetGoalCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "set-goal <track-name> <goal-slug>",
		Short: "Link a track to a goal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			trackName, goalSlug := args[0], args[1]

			t, err := s.GetTrack(trackName)
			if err != nil {
				return err
			}
			if _, err := s.LoadGoal(goalSlug); err != nil {
				return fmt.Errorf("goal %q not found", goalSlug)
			}

			t.GoalSlug = goalSlug
			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Linked %s → goal %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(trackName), ui.Highlight.Render(goalSlug))
			return nil
		},
	}
}

func newTrackDependsOnCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "depends-on <track-name> <dependency-track>",
		Short: "Add a prerequisite track",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			trackName, dep := args[0], args[1]

			t, err := s.GetTrack(trackName)
			if err != nil {
				return err
			}
			if _, err := s.GetTrack(dep); err != nil {
				return fmt.Errorf("dependency track %q not found", dep)
			}

			for _, existing := range t.DependsOn {
				if existing == dep {
					fmt.Printf("%s already depends on %s\n", trackName, dep)
					return nil
				}
			}

			t.DependsOn = append(t.DependsOn, dep)

			// Load all tracks to check for cycles
			allTracks, err := s.ListTracks()
			if err != nil {
				return err
			}
			// Replace the in-memory track with updated version for cycle check
			for i, tt := range allTracks {
				if tt.Name == trackName {
					allTracks[i] = t
				}
			}
			if model.HasCycle(allTracks) {
				return fmt.Errorf("adding dependency %q → %q would create a cycle", trackName, dep)
			}

			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s %s now depends on %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(trackName), ui.Highlight.Render(dep))
			return nil
		},
	}
}

func newTrackMilestoneCmd(getStore func() *store.FSStore) *cobra.Command {
	ms := &cobra.Command{
		Use:   "milestone",
		Short: "Manage track milestones",
	}
	ms.AddCommand(
		newTrackMilestoneAddCmd(getStore),
		newTrackMilestoneDoneCmd(getStore),
	)
	return ms
}

func newTrackMilestoneAddCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		description string
		deadline    string
		artifact    string
	)
	cmd := &cobra.Command{
		Use:   "add <track-name>",
		Short: "Add a milestone to a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			t, err := s.GetTrack(args[0])
			if err != nil {
				return err
			}

			if description == "" {
				description = mustPrompt("Milestone description", "", func(v string) error {
					if strings.TrimSpace(v) == "" {
						return fmt.Errorf("description cannot be empty")
					}
					return nil
				})
			}

			m := model.Milestone{
				ID:               store.Slugify(description),
				Description:      description,
				ArtifactResource: artifact,
			}
			if deadline != "" {
				d, err := time.Parse("2006-01-02", deadline)
				if err != nil {
					return fmt.Errorf("deadline must be YYYY-MM-DD, got %q", deadline)
				}
				m.Deadline = &d
			}

			t.Milestones = append(t.Milestones, m)
			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Added milestone to %s: %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(t.Name), m.Description)
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "milestone description")
	cmd.Flags().StringVar(&deadline, "deadline", "", "target date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}

func newTrackMilestoneDoneCmd(getStore func() *store.FSStore) *cobra.Command {
	var artifact string
	cmd := &cobra.Command{
		Use:   "done <track-name> <milestone-id>",
		Short: "Mark a track milestone complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			t, err := s.GetTrack(args[0])
			if err != nil {
				return err
			}

			found := false
			now := time.Now()
			for i, m := range t.Milestones {
				if m.ID == args[1] {
					t.Milestones[i].CompletedAt = &now
					if artifact != "" {
						t.Milestones[i].ArtifactResource = artifact
					}
					found = true
					break
				}
			}
			if !found {
				ids := make([]string, len(t.Milestones))
				for i, m := range t.Milestones {
					ids[i] = m.ID
				}
				return fmt.Errorf("milestone %q not found; available: %v", args[1], ids)
			}

			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Milestone complete: %s\n", ui.Success.Render("✓"), args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}
```

Note: `track.go` will need these imports added: `"strings"`, `"time"`, `"github.com/TTitcombe/questlog/internal/model"`.

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 10: Add --core flag to add and classify

**Files:**
- Modify: `internal/cli/add.go`
- Modify: `internal/cli/classify.go`

- [ ] **Step 1: Add `--core` flag to `newAddCmd` in `add.go`**

Add `isCore bool` to the var block at the top of `newAddCmd`:

```go
var (
    // ... existing vars ...
    isCore   bool
)
```

Add the flag registration at the bottom of `newAddCmd`:

```go
cmd.Flags().BoolVar(&isCore, "core", false, "mark as a core (non-negotiable) resource for its track")
```

In the `RunE`, set `IsCore` on the resource before saving. Replace the `r := model.Resource{...}` block with:

```go
r := model.Resource{
    Title:            title,
    Type:             rt,
    URL:              url,
    Tags:             tagList,
    Track:            track,
    Added:            time.Now(),
    EstimatedMinutes: minutes,
    Status:           model.StatusUnread,
    Priority:         priority,
    IsCore:           isCore,
}
```

The quick-capture path (`if quick != ""`) does not support `--core` (inbox items are never core); leave it unchanged.

- [ ] **Step 2: Add `--core` flag to `newClassifyCmd` in `classify.go`**

Add `isCore bool` and `setCore bool` to the var block:

```go
var (
    track   string
    isCore  bool
)
```

Add the flag:

```go
cmd.Flags().BoolVar(&isCore, "core", false, "mark as a core resource after moving")
```

After the `MoveToTrack` call succeeds, if `--core` was set, load and re-save the resource:

```go
if err := s.MoveToTrack(id, track); err != nil {
    return err
}
if isCore {
    r, err := s.GetResource(id)
    if err != nil {
        return err
    }
    r.IsCore = true
    if err := s.SaveResource(r); err != nil {
        return err
    }
}
fmt.Printf("%s Moved %s → %s\n", ui.Success.Render("✓"), ui.Dim.Render(id), ui.Highlight.Render(track))
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

---

## Task 11: Goal-aware focus

**Files:**
- Modify: `internal/cli/focus_tui.go`
- Modify: `internal/cli/focus.go`

- [ ] **Step 1: Add goalContext type and field to focusModel in `focus_tui.go`**

Add the `goalContext` struct and add a `gc *goalContext` field to `focusModel`, after the `notice` field:

```go
type goalContext struct {
	goalTitle         string
	pendingMilestones []string // formatted strings, both goal and track level
	availableTracks   []string
	tracksComplete    int
	tracksTotal       int
}
```

Add to `focusModel`:

```go
gc *goalContext // nil when no active goal
```

- [ ] **Step 2: Render goal header in `View()` in `focus_tui.go`**

In the `View()` method, replace the header block:

```go
// Header
timerStr := formatFocusDuration(m.remaining)
if m.state == focusStateTimesUp {
    timerStr = ui.Warning.Render("Time's up!")
}
b.WriteString(ui.Bold.Render("Focus session") + "  " + timerStr + "\n\n")
```

with:

```go
// Header
timerStr := formatFocusDuration(m.remaining)
if m.state == focusStateTimesUp {
    timerStr = ui.Warning.Render("Time's up!")
}

if m.gc != nil {
    pct := 0
    if m.gc.tracksTotal > 0 {
        pct = m.gc.tracksComplete * 100 / m.gc.tracksTotal
    }
    b.WriteString(ui.Bold.Render("Goal: "+m.gc.goalTitle) +
        "  " + ui.Dim.Render(fmt.Sprintf("[%d%%]", pct)) + "\n")
    if len(m.gc.availableTracks) > 0 {
        b.WriteString(ui.Muted.Render("Available: "+strings.Join(m.gc.availableTracks, ", ")) + "\n")
    }
    for _, ms := range m.gc.pendingMilestones {
        b.WriteString(ui.Warning.Render("⚑ "+ms) + "\n")
    }
    b.WriteString("\n")
}

b.WriteString(ui.Bold.Render("Focus session") + "  " + timerStr + "\n\n")
```

- [ ] **Step 3: Update `newFocusModel` signature to accept `*goalContext`**

```go
func newFocusModel(session, noEst, all []model.Resource, minutes int, s *store.FSStore, gc *goalContext) focusModel {
    return focusModel{
        session:    session,
        noEst:      noEst,
        browseList: all,
        remaining:  time.Duration(minutes) * time.Minute,
        store:      s,
        startedAt:  time.Now(),
        gc:         gc,
    }
}
```

- [ ] **Step 4: Build to catch signature mismatch**

```bash
go build ./...
```

Expected: compile error in `focus.go` — wrong number of args to `newFocusModel`. That's expected; fix it in the next step.

- [ ] **Step 5: Update `focus.go` to build and pass goalContext**

In `newFocusCmd`, before the `newFocusModel` call, build the `goalContext`. Replace:

```go
m := newFocusModel(session, noEst, candidates, minutes, s)
```

with:

```go
gc := buildGoalContext(s, track)
m := newFocusModel(session, noEst, candidates, minutes, s, gc)
```

Then add `buildGoalContext` to `focus.go`:

```go
// buildGoalContext returns a goalContext if there are active goals and no
// specific track was requested. Returns nil otherwise.
func buildGoalContext(s *store.FSStore, trackFlag string) *goalContext {
    if trackFlag != "" {
        return nil
    }
    goals, err := s.ListGoals()
    if err != nil || len(goals) == 0 {
        return nil
    }

    // Pick the goal with the nearest pending milestone deadline.
    var picked *model.Goal
    for i := range goals {
        g := &goals[i]
        for _, m := range g.Milestones {
            if m.CompletedAt == nil && m.Deadline != nil {
                if picked == nil {
                    picked = g
                } else {
                    // find nearest deadline in picked
                    for _, pm := range picked.Milestones {
                        if pm.CompletedAt == nil && pm.Deadline != nil {
                            if m.Deadline.Before(*pm.Deadline) {
                                picked = g
                            }
                            break
                        }
                    }
                }
                break
            }
        }
    }
    if picked == nil {
        picked = &goals[0]
    }

    allTracks, _ := s.ListTracks()
    var goalTracks []model.Track
    for _, t := range allTracks {
        if t.GoalSlug == picked.Slug {
            goalTracks = append(goalTracks, t)
        }
    }

    resources := make(map[string][]model.Resource, len(goalTracks))
    for _, t := range goalTracks {
        res, _ := s.ListResources(store.ResourceFilter{Track: t.Name})
        resources[t.Name] = res
    }

    gp := model.ComputeGoalProgress(*picked, goalTracks, resources)

    var available []string
    var pending []string
    for _, tp := range gp.Tracks {
        if tp.Status == model.TrackStatusAvailable {
            available = append(available, tp.Track.Name)
            for _, m := range tp.Track.Milestones {
                if m.CompletedAt == nil && m.Deadline != nil {
                    pending = append(pending, fmt.Sprintf("%s · due %s [%s]",
                        m.Description, m.Deadline.Format("Jan 2"), tp.Track.Name))
                    break
                }
            }
        }
    }
    for _, m := range picked.Milestones {
        if m.CompletedAt == nil && m.Deadline != nil {
            pending = append(pending, fmt.Sprintf("%s · due %s [goal]",
                m.Description, m.Deadline.Format("Jan 2")))
            break
        }
    }

    return &goalContext{
        goalTitle:         picked.Title,
        pendingMilestones: pending,
        availableTracks:   available,
        tracksComplete:    gp.TracksComplete,
        tracksTotal:       gp.TracksTotal,
    }
}
```

Also update the resource filter in focus.go: when `gc` is non-nil and `trackFlag` is empty, filter resources to available tracks only:

In `newFocusCmd`'s `RunE`, after `gc := buildGoalContext(...)`, replace:

```go
filter := store.ResourceFilter{Track: track}
all, err := s.ListResources(filter)
```

with:

```go
var all []model.Resource
if gc != nil && len(gc.availableTracks) > 0 {
    for _, name := range gc.availableTracks {
        res, err := s.ListResources(store.ResourceFilter{Track: name})
        if err != nil {
            return err
        }
        all = append(all, res...)
    }
} else {
    all, err = s.ListResources(store.ResourceFilter{Track: track})
    if err != nil {
        return err
    }
}
```

Note: `focus.go` will need `"fmt"` in its import block (it may already be there).

- [ ] **Step 6: Build**

```bash
go build ./...
```

Expected: no output, exit 0.

- [ ] **Step 7: Smoke test the full binary**

```bash
make build
bin/qlog goal new "Test Goal" --description "A test"
bin/qlog goal list
bin/qlog track new test-track --goal test-goal
bin/qlog progress test-goal
```

Expected:
- `goal new` prints "✓ Created goal: Test Goal"
- `goal list` shows the goal with 1 track
- `track new` prints "✓ Created track: test-track"  
- `progress test-goal` shows `test-track` with "no core resources set" warning

```bash
bin/qlog add --title "Key Paper" --type paper --track test-track --core
bin/qlog progress test-goal
```

Expected: `test-track` now shows `0/1 core`.

```bash
bin/qlog done key-paper
bin/qlog progress test-goal
```

Expected: `test-track` still shows `◑` (0/1 milestones). It requires at least one milestone to be complete.

```bash
bin/qlog track milestone add test-track --description "Can explain the concept" 
bin/qlog track milestone done test-track can-explain-the-concept
bin/qlog progress test-goal
```

Expected: `test-track` shows `✓` (complete).

---

## Self-Review Checklist

- [x] **Spec coverage:** Goal CRUD ✓, Track extensions (goal_slug, depends_on, milestones) ✓, Resource is_core ✓, Milestone struct ✓, Progress computation ✓, DAG cycle detection ✓, CLI commands (goal new/list/show/milestone add/done, progress, track set-goal/depends-on/milestone add/done, add --core, classify --core) ✓, Focus goal-awareness ✓
- [x] **Placeholders:** None — all steps have concrete code
- [x] **Type consistency:** `model.Milestone` used in Goal and Track. `store.Slugify` exported in Task 2, used in Tasks 6 and 9. `model.HasCycle` defined in Task 5, used in Task 9. `model.ComputeGoalProgress` defined in Task 5, used in Tasks 7 and 11. `goalContext` defined in Task 11 Step 1 (`focus_tui.go`), used in Task 11 Steps 3–5 (`focus.go`). `newFocusModel` signature updated in Step 3, callers updated in Step 5.
