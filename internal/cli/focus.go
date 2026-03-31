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

			filter := store.ResourceFilter{Track: track}
			all, err := s.ListResources(filter)
			if err != nil {
				return err
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

			m := newFocusModel(session, noEst, minutes, s)
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
