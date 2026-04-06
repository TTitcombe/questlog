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
		track    string
		rtype    string
		inclNotes bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search resources by title, tags, and track",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			query := strings.Join(args, " ")

			// Collect results: index search + optional notes scan, deduplicated by ID.
			seen := map[string]bool{}
			type result struct {
				id, title, track, status, rtype string
				mins                            int
			}
			var out []result

			indexResults, err := s.SearchIndex(query)
			if err != nil {
				return err
			}
			for _, e := range indexResults {
				if track != "" && e.Track != track {
					continue
				}
				if rtype != "" && string(e.Type) != rtype {
					continue
				}
				if !seen[e.ID] {
					seen[e.ID] = true
					out = append(out, result{e.ID, e.Title, e.Track, string(e.Status), string(e.Type), e.EstimatedMinutes})
				}
			}

			if inclNotes {
				noteResults, err := s.SearchNotes(query)
				if err != nil {
					return err
				}
				for _, r := range noteResults {
					if track != "" && r.Track != track {
						continue
					}
					if rtype != "" && string(r.Type) != rtype {
						continue
					}
					if !seen[r.ID] {
						seen[r.ID] = true
						out = append(out, result{r.ID, r.Title, r.Track, string(r.Status), string(r.Type), r.EstimatedMinutes})
					}
				}
			}

			if len(out) == 0 {
				fmt.Printf("%s No results for %q\n", ui.Warning.Render("!"), query)
				return nil
			}

			fmt.Printf("%s results for %q\n\n", ui.Highlight.Render(fmt.Sprintf("%d", len(out))), query)
			for _, e := range out {
				mins := ""
				if e.mins > 0 {
					mins = fmt.Sprintf("  %s", ui.Dim.Render(fmt.Sprintf("~%dm", e.mins)))
				}
				line := fmt.Sprintf("  %s  %s  %s  %s%s\n     %s",
					ui.StatusBadge(e.status),
					ui.TypeBadge(e.rtype),
					ui.Dim.Render("["+e.track+"]"),
					e.title,
					mins,
					ui.Dim.Render(e.id),
				)
				fmt.Println(line)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "filter results by track")
	cmd.Flags().StringVar(&rtype, "type", "", "filter results by type")
	cmd.Flags().BoolVar(&inclNotes, "notes", false, "also search resource note bodies (slower)")
	return cmd
}
