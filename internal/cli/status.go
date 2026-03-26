package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newStatusCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show an overview of your questlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			idx, err := s.GetIndex()
			if err != nil {
				return err
			}

			total := len(idx.Entries)
			done, inProgress, unread, inboxCount := 0, 0, 0, 0
			trackMap := map[string]struct{ total, done int }{}

			for _, e := range idx.Entries {
				switch e.Status {
				case model.StatusDone:
					done++
				case model.StatusInProgress:
					inProgress++
				default:
					unread++
				}
				if e.Track == "inbox" {
					inboxCount++
					continue
				}
				stats := trackMap[e.Track]
				stats.total++
				if e.Status == model.StatusDone {
					stats.done++
				}
				trackMap[e.Track] = stats
			}

			fmt.Printf("%s\n\n", ui.Bold.Render("questlog status"))

			pct := 0
			if total > 0 {
				pct = done * 100 / total
			}
			fmt.Printf("Overall  %s  %d/%d complete (%d%%)\n",
				ui.ProgressBar(pct, 20), done, total, pct)
			fmt.Printf("         %s in-progress  %s unread  %s inbox\n\n",
				ui.Warning.Render(fmt.Sprintf("%d", inProgress)),
				ui.Muted.Render(fmt.Sprintf("%d", unread)),
				ui.Muted.Render(fmt.Sprintf("%d", inboxCount)),
			)

			if len(trackMap) > 0 {
				fmt.Println(ui.Bold.Render("Tracks:"))
				tracks, _ := s.ListTracks()
				for _, t := range tracks {
					stats := trackMap[t.Name]
					tp := 0
					if stats.total > 0 {
						tp = stats.done * 100 / stats.total
					}
					fmt.Printf("  %-20s %s %d/%d\n",
						ui.Highlight.Render(t.Name),
						ui.ProgressBar(tp, 12),
						stats.done, stats.total,
					)
				}
			}

			return nil
		},
	}
}
