package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/store"
)

func newNoteCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "note <id> <text>",
		Short: "Append a note to a resource",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			r, err := s.GetResource(args[0])
			if err != nil {
				return err
			}
			text := strings.Join(args[1:], " ")
			r.Notes = prependNote(r.Notes, text, time.Now())
			if err := s.SaveResource(r); err != nil {
				return err
			}
			fmt.Printf("Note added to %q\n", r.Title)
			return nil
		},
	}
}

// prependNote prepends a timestamped entry to the notes body.
// Newer entries appear at the top. Used by both the CLI command and focus TUI.
func prependNote(existing, text string, t time.Time) string {
	entry := fmt.Sprintf("<!-- %s -->\n%s\n", t.Format("2006-01-02 15:04"), strings.TrimSpace(text))
	if existing == "" {
		return entry
	}
	return entry + "\n" + existing
}
