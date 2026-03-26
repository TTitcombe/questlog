package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newTrackCmd(getStore func() *store.FSStore) *cobra.Command {
	track := &cobra.Command{
		Use:   "track",
		Short: "Manage learning tracks",
	}
	track.AddCommand(
		newTrackNewCmd(getStore),
		newTrackListCmd(getStore),
		newTrackShowCmd(getStore),
	)
	return track
}

func newTrackNewCmd(getStore func() *store.FSStore) *cobra.Command {
	var description string
	var tags string

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new learning track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			s := getStore()

			// Check it doesn't already exist
			if _, err := s.GetTrack(name); err == nil {
				return fmt.Errorf("track %q already exists", name)
			}

			var tagList []string
			for _, t := range strings.Split(tags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagList = append(tagList, t)
				}
			}

			t := model.Track{
				Name:        name,
				Description: description,
				Tags:        tagList,
				Created:     time.Now(),
			}
			if err := s.CreateTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Created track: %s\n", ui.Success.Render("✓"), ui.Highlight.Render(name))
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "track description")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	return cmd
}

func newTrackListCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all tracks",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			tracks, err := s.ListTracks()
			if err != nil {
				return err
			}
			if len(tracks) == 0 {
				fmt.Println(ui.Muted.Render("No tracks yet. Create one with: qlog track new <name>"))
				return nil
			}

			idx, _ := s.GetIndex()

			for _, t := range tracks {
				// Count resources in this track
				total, done := 0, 0
				for _, e := range idx.Entries {
					if e.Track == t.Name {
						total++
						if e.Status == model.StatusDone {
							done++
						}
					}
				}
				pct := 0
				if total > 0 {
					pct = done * 100 / total
				}

				fmt.Printf("  %s", ui.Highlight.Render(t.Name))
				if t.Description != "" {
					fmt.Printf("  %s", ui.Muted.Render(t.Description))
				}
				fmt.Println()
				fmt.Printf("    %s  %d/%d complete  %s\n",
					ui.ProgressBar(pct, 15),
					done, total,
					ui.Dim.Render(fmt.Sprintf("%d%%", pct)),
				)
			}
			return nil
		},
	}
}

func newTrackShowCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details of a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			s := getStore()

			t, err := s.GetTrack(name)
			if err != nil {
				return err
			}

			resources, err := s.ListResources(store.ResourceFilter{Track: name})
			if err != nil {
				return err
			}

			done := 0
			for _, r := range resources {
				if r.Status == model.StatusDone {
					done++
				}
			}
			total := len(resources)
			pct := 0
			if total > 0 {
				pct = done * 100 / total
			}

			fmt.Printf("%s\n", ui.Bold.Render(t.Name))
			if t.Description != "" {
				fmt.Printf("%s\n", ui.Muted.Render(t.Description))
			}
			fmt.Printf("Progress: %s %d/%d (%d%%)\n\n",
				ui.ProgressBar(pct, 20), done, total, pct)

			if len(resources) == 0 {
				fmt.Println(ui.Muted.Render("No resources yet. Add some with: qlog add --track " + name))
				return nil
			}

			for _, r := range resources {
				mins := ""
				if r.EstimatedMinutes > 0 {
					mins = fmt.Sprintf("  %s", ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes)))
				}
				fmt.Printf("  %s  %s  %s%s\n",
					ui.StatusBadge(string(r.Status)),
					ui.TypeBadge(string(r.Type)),
					r.Title,
					mins,
				)
				fmt.Printf("     %s\n", ui.Dim.Render(r.ID))
			}
			return nil
		},
	}
}
