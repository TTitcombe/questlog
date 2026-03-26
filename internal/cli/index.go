package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/store"
)

func newIndexCmd(getStore func() *store.FSStore) *cobra.Command {
	index := &cobra.Command{
		Use:    "index",
		Short:  "Manage the search index",
		Hidden: true,
	}

	index.AddCommand(&cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild index.json from all markdown files",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			if err := s.RebuildIndex(); err != nil {
				return err
			}
			fmt.Println(ui.Success.Render("✓ Index rebuilt"))
			return nil
		},
	})

	return index
}
