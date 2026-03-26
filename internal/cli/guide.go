package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newGuideCmd(getStore func() *store.FSStore) *cobra.Command {
	var track string

	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Generate a prioritized reading guide for a track",
		Example: `  qlog guide --track llm
  qlog guide  # shows guide across all tracks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()

			filter := store.ResourceFilter{Track: track}
			all, err := s.ListResources(filter)
			if err != nil {
				return err
			}

			// Exclude done
			var resources []model.Resource
			for _, r := range all {
				if r.Status != model.StatusDone {
					resources = append(resources, r)
				}
			}

			if len(resources) == 0 {
				if track != "" {
					fmt.Printf("No incomplete resources in track %q.\n", track)
				} else {
					fmt.Println("No incomplete resources found.")
				}
				return nil
			}

			// Sort: in-progress first, then by priority (1 highest; 0/unset last), then by estimated_minutes ASC
			sort.SliceStable(resources, func(i, j int) bool {
				ri, rj := resources[i], resources[j]
				// in-progress before unread
				if ri.Status != rj.Status {
					if ri.Status == model.StatusInProgress {
						return true
					}
					if rj.Status == model.StatusInProgress {
						return false
					}
				}
				// priority: lower number = higher priority; 0 = unset, goes last
				pi, pj := ri.Priority, rj.Priority
				if pi == 0 {
					pi = 6
				}
				if pj == 0 {
					pj = 6
				}
				if pi != pj {
					return pi < pj
				}
				// tiebreak by estimated_minutes ASC
				return ri.EstimatedMinutes < rj.EstimatedMinutes
			})

			header := "Reading guide"
			if track != "" {
				header += " — " + track
			}
			fmt.Printf("%s\n\n", ui.Bold.Render(header))

			// Group by track if showing all
			if track == "" {
				byTrack := map[string][]model.Resource{}
				var order []string
				for _, r := range resources {
					if _, seen := byTrack[r.Track]; !seen {
						order = append(order, r.Track)
					}
					byTrack[r.Track] = append(byTrack[r.Track], r)
				}
				for _, t := range order {
					fmt.Printf("%s\n", ui.Highlight.Render(t+":"))
					printGuideList(byTrack[t])
					fmt.Println()
				}
			} else {
				printGuideList(resources)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "track to generate guide for (default: all)")
	return cmd
}

func printGuideList(resources []model.Resource) {
	for i, r := range resources {
		// Priority badge
		pri := ""
		if r.Priority > 0 {
			stars := strings.Repeat("★", 6-r.Priority) + strings.Repeat("☆", r.Priority-1)
			pri = ui.Warning.Render(stars) + " "
		}

		mins := ""
		if r.EstimatedMinutes > 0 {
			mins = ui.Dim.Render(fmt.Sprintf(" ~%dm", r.EstimatedMinutes))
		}

		progress := ""
		if r.Status == model.StatusInProgress && r.Progress > 0 {
			progress = ui.Warning.Render(fmt.Sprintf(" [%d%%]", r.Progress))
		}

		fmt.Printf("  %s%s %s%s%s%s\n",
			ui.Dim.Render(fmt.Sprintf("%2d.", i+1)),
			" "+pri,
			ui.TypeBadge(string(r.Type)),
			" "+r.Title,
			mins,
			progress,
		)
		if r.URL != "" {
			fmt.Printf("      %s\n", ui.Muted.Render(r.URL))
		}
		fmt.Printf("      %s\n", ui.Dim.Render(r.ID))
	}
}
