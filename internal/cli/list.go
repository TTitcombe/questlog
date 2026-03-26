package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newListCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		track  string
		status string
		rtype  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Example: `  qlog list
  qlog list --track llm --status unread
  qlog list --type paper`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			filter := store.ResourceFilter{
				Track:  track,
				Status: model.Status(status),
				Type:   model.ResourceType(rtype),
			}
			resources, err := s.ListResources(filter)
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				fmt.Println(ui.Muted.Render("No resources found."))
				return nil
			}

			for _, r := range resources {
				mins := ""
				if r.EstimatedMinutes > 0 {
					mins = fmt.Sprintf("  %s", ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes)))
				}
				fmt.Printf("  %s  %s  %s  %s%s\n",
					ui.StatusBadge(string(r.Status)),
					ui.TypeBadge(string(r.Type)),
					ui.Dim.Render("["+r.Track+"]"),
					r.Title,
					mins,
				)
				fmt.Printf("     %s\n", ui.Dim.Render(r.ID))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "filter by track")
	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status (unread|in-progress|done)")
	cmd.Flags().StringVar(&rtype, "type", "", "filter by type (paper|video|book|article|note|idea)")
	return cmd
}
