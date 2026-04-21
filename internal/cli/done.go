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

