package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/store"
)

func newSearchCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		track string
		rtype string
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search resources by title, tags, and track",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			query := strings.Join(args, " ")

			results, err := s.SearchIndex(query)
			if err != nil {
				return err
			}

			// Filter by track/type flags
			var filtered []interface{ GetTrack() string }
			_ = filtered // unused; inline filtering below

			var out []string
			for _, e := range results {
				if track != "" && e.Track != track {
					continue
				}
				if rtype != "" && string(e.Type) != rtype {
					continue
				}
				mins := ""
				if e.EstimatedMinutes > 0 {
					mins = fmt.Sprintf("  %s", ui.Dim.Render(fmt.Sprintf("~%dm", e.EstimatedMinutes)))
				}
				line := fmt.Sprintf("  %s  %s  %s  %s%s\n     %s",
					ui.StatusBadge(string(e.Status)),
					ui.TypeBadge(string(e.Type)),
					ui.Dim.Render("["+e.Track+"]"),
					e.Title,
					mins,
					ui.Dim.Render(e.ID),
				)
				out = append(out, line)
			}

			if len(out) == 0 {
				fmt.Printf("%s No results for %q\n", ui.Warning.Render("!"), query)
				return nil
			}

			fmt.Printf("%s results for %q\n\n", ui.Highlight.Render(fmt.Sprintf("%d", len(out))), query)
			for _, line := range out {
				fmt.Println(line)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "filter results by track")
	cmd.Flags().StringVar(&rtype, "type", "", "filter results by type")
	return cmd
}
