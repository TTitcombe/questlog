package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
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
		Short: "Suggest resources for a focus session",
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

			// Split into tiers
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

			// Greedy bin-pack
			var session []model.Resource
			remaining := minutes
			for _, r := range append(tier1, tier2...) {
				if r.EstimatedMinutes <= remaining {
					session = append(session, r)
					remaining -= r.EstimatedMinutes
				}
			}

			header := fmt.Sprintf("Focus session — %d minutes", minutes)
			if track != "" {
				header += " · " + track
			}
			fmt.Println(ui.Bold.Render(header))
			fmt.Println()

			if len(session) == 0 {
				// Nothing fits — show closest fit
				fmt.Println(ui.Warning.Render("Nothing fits in the available time. Closest option:"))
				closest := tier1
				closest = append(closest, tier2...)
				if len(closest) > 0 {
					r := closest[0]
					fmt.Printf("  %s  %s  %s  (~%dm)\n",
						ui.StatusBadge(string(r.Status)),
						ui.TypeBadge(string(r.Type)),
						ui.Bold.Render(r.Title),
						r.EstimatedMinutes,
					)
				}
			} else {
				used := minutes - remaining
				fmt.Printf("Suggested (%d/%d min used):\n\n", used, minutes)
				for _, r := range session {
					fmt.Printf("  %s  %s  %s  %s\n",
						ui.StatusBadge(string(r.Status)),
						ui.TypeBadge(string(r.Type)),
						ui.Bold.Render(r.Title),
						ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes)),
					)
				}
			}

			if len(noEst) > 0 {
				fmt.Printf("\n%s\n", ui.Muted.Render("Also consider (no time estimate):"))
				for _, r := range noEst {
					fmt.Printf("  %s  %s\n", ui.TypeBadge(string(r.Type)), r.Title)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "limit to a specific track")
	cmd.Flags().IntVarP(&minutes, "minutes", "m", 30, "available time in minutes")
	return cmd
}
