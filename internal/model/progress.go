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
