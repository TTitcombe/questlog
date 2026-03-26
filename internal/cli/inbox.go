package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/store"
)

func newInboxCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "inbox",
		Short: "List uncategorised inbox items",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			resources, err := s.ListInbox()
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				fmt.Println(ui.Muted.Render("Inbox is empty. Add something with: qlog add --quick \"idea\""))
				return nil
			}
			fmt.Printf("%s  %d item(s)\n\n", ui.Highlight.Render("Inbox"), len(resources))
			for _, r := range resources {
				fmt.Printf("  %s  %s  %s\n",
					ui.Dim.Render(r.ID),
					ui.TypeBadge(string(r.Type)),
					ui.Bold.Render(r.Title),
				)
				if r.URL != "" {
					fmt.Printf("        %s\n", ui.Muted.Render(r.URL))
				}
			}
			fmt.Printf("\nUse %s to move an item to a track.\n", ui.Dim.Render("qlog classify <id> --track <name>"))
			return nil
		},
	}
}
