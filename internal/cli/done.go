package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newDoneCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a resource as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			r, err := s.GetResource(args[0])
			if err != nil {
				return err
			}
			r.Status = model.StatusDone
			r.Progress = 100
			if err := s.SaveResource(r); err != nil {
				return err
			}
			fmt.Printf("%s Completed: %s\n", ui.Success.Render("✓"), ui.Bold.Render(r.Title))
			return nil
		},
	}
}

func newProgressCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "progress <id> <0-100>",
		Short: "Set progress on a resource",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			r, err := s.GetResource(args[0])
			if err != nil {
				return err
			}
			pct := 0
			if _, err := fmt.Sscanf(args[1], "%d", &pct); err != nil || pct < 0 || pct > 100 {
				return fmt.Errorf("progress must be an integer between 0 and 100")
			}
			r.Progress = pct
			if pct == 100 {
				r.Status = model.StatusDone
			} else if pct > 0 {
				r.Status = model.StatusInProgress
			}
			if err := s.SaveResource(r); err != nil {
				return err
			}
			fmt.Printf("%s %s → %s %d%%\n",
				ui.Success.Render("✓"),
				r.Title,
				ui.ProgressBar(pct, 15),
				pct,
			)
			return nil
		},
	}
}
