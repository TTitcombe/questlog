package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/store"
)

func newClassifyCmd(getStore func() *store.FSStore) *cobra.Command {
	var track string

	cmd := &cobra.Command{
		Use:   "classify <id>",
		Short: "Move an inbox item to a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			id := args[0]

			if track == "" {
				tracks, err := s.ListTracks()
				if err != nil || len(tracks) == 0 {
					return fmt.Errorf("no tracks exist; create one with: qlog track new <name>")
				}
				names := make([]string, len(tracks))
				for i, t := range tracks {
					names[i] = t.Name
				}
				track = mustSelect("Move to track", names)
			}

			if err := s.MoveToTrack(id, track); err != nil {
				return err
			}
			fmt.Printf("%s Moved %s → %s\n", ui.Success.Render("✓"), ui.Dim.Render(id), ui.Highlight.Render(track))
			return nil
		},
	}

	cmd.Flags().StringVarP(&track, "track", "t", "", "target track name")
	return cmd
}
