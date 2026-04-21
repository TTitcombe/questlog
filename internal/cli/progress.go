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
