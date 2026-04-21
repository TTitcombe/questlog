package cli

import (
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

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

func newFocusCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		track   string
		minutes int
	)

	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Start an interactive focus session",
		Example: `  qlog focus --minutes 30
  qlog focus --track llm --minutes 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()

			gc := buildGoalContext(s, track)

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
				var err error
				all, err = s.ListResources(store.ResourceFilter{Track: track})
				if err != nil {
					return err
				}
			}

			// Exclude done resources
			var candidates []model.Resource
			for _, r := range all {
				if r.Status != model.StatusDone {
					candidates = append(candidates, r)
				}
			}

			if len(candidates) == 0 {
				if track != "" {
					fmt.Printf("No incomplete resources in track %q.\n", track)
				} else {
					fmt.Println("No incomplete resources found.")
				}
				return nil
			}

			// Split into tiers: in-progress first, then others, then no estimate
			var tier1, tier2, noEst []model.Resource
			for _, r := range candidates {
				if r.EstimatedMinutes == 0 {
					noEst = append(noEst, r)
				} else if r.Status == model.StatusInProgress {
					tier1 = append(tier1, r)
				} else {
					tier2 = append(tier2, r)
				}
			}

			byMins := func(a, b model.Resource) bool {
				return a.EstimatedMinutes < b.EstimatedMinutes
			}
			sort.Slice(tier1, func(i, j int) bool { return byMins(tier1[i], tier1[j]) })
			sort.Slice(tier2, func(i, j int) bool { return byMins(tier2[i], tier2[j]) })

			// Greedy bin-pack into session
			var session []model.Resource
			remaining := minutes
			for _, r := range append(tier1, tier2...) {
				if r.EstimatedMinutes <= remaining {
					session = append(session, r)
					remaining -= r.EstimatedMinutes
				}
			}

			m := newFocusModel(session, noEst, candidates, minutes, s, gc)
			p := tea.NewProgram(m, tea.WithAltScreen())
			finalModel, err := p.Run()
			if err != nil {
				return err
			}

			if fm, ok := finalModel.(focusModel); ok {
				actualSecs := int(time.Since(fm.startedAt).Seconds())
				if len(fm.opened) > 0 || len(fm.statusChanges) > 0 || actualSecs >= 60 {
					sess := model.Session{
						StartedAt:     fm.startedAt,
						PlannedMins:   minutes,
						ActualSecs:    actualSecs,
						Track:         track,
						Opened:        fm.opened,
						StatusChanges: fm.statusChanges,
					}
					_ = s.AppendSession(sess)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "limit to a specific track")
	cmd.Flags().IntVarP(&minutes, "minutes", "m", 30, "available time in minutes")
	return cmd
}
