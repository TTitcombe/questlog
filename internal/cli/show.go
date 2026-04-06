package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/store"
)

func newShowCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a resource's details and notes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			r, err := s.GetResource(args[0])
			if err != nil {
				return err
			}

			fmt.Println(ui.Bold.Render(r.Title))

			meta := fmt.Sprintf("  %s  %s  %s",
				ui.TypeBadge(string(r.Type)),
				ui.Dim.Render("["+r.Track+"]"),
				ui.StatusBadge(string(r.Status)),
			)
			if r.Rating != nil {
				meta += "  " + formatRating(r.Rating)
			}
			fmt.Println(meta)

			if r.URL != "" {
				fmt.Println("  " + ui.Muted.Render(r.URL))
			}

			extras := []string{}
			if r.EstimatedMinutes > 0 {
				extras = append(extras, fmt.Sprintf("~%dm", r.EstimatedMinutes))
			}
			if r.Priority > 0 {
				extras = append(extras, fmt.Sprintf("priority %d", r.Priority))
			}
			if !r.Added.IsZero() {
				extras = append(extras, "added "+r.Added.Format("2006-01-02"))
			}
			if len(extras) > 0 {
				fmt.Println("  " + ui.Dim.Render(strings.Join(extras, " · ")))
			}

			if len(r.Tags) > 0 {
				fmt.Println("  " + ui.Dim.Render("tags: "+strings.Join(r.Tags, ", ")))
			}

			fmt.Println()
			fmt.Println(ui.Bold.Render("Notes"))
			fmt.Println(strings.Repeat("─", 30))
			if r.Notes == "" {
				fmt.Println(ui.Muted.Render("no notes yet"))
			} else {
				fmt.Print(r.Notes)
				if !strings.HasSuffix(r.Notes, "\n") {
					fmt.Println()
				}
			}

			return nil
		},
	}
}

func formatRating(r *int) string {
	if r == nil {
		return ui.Muted.Render("unrated")
	}
	switch *r {
	case 1:
		return ui.Success.Render("+1 valuable")
	case -1:
		return ui.Warning.Render("-1 not worth it")
	default:
		return ui.Dim.Render("0 neutral")
	}
}
