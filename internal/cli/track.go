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
		newTrackSetGoalCmd(getStore),
		newTrackDependsOnCmd(getStore),
		newTrackMilestoneCmd(getStore),
	)
	return track
}

func newTrackNewCmd(getStore func() *store.FSStore) *cobra.Command {
	var description string
	var tags string
	var goalSlug string

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

			if goalSlug != "" {
				if _, err := s.LoadGoal(goalSlug); err != nil {
					return fmt.Errorf("goal %q not found; create it first with: qlog goal new", goalSlug)
				}
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
				GoalSlug:    goalSlug,
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
	cmd.Flags().StringVar(&goalSlug, "goal", "", "link this track to a goal (goal slug)")
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

func newTrackSetGoalCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "set-goal <track-name> <goal-slug>",
		Short: "Link a track to a goal",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			trackName, goalSlug := args[0], args[1]

			t, err := s.GetTrack(trackName)
			if err != nil {
				return err
			}
			if _, err := s.LoadGoal(goalSlug); err != nil {
				return fmt.Errorf("goal %q not found", goalSlug)
			}

			t.GoalSlug = goalSlug
			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Linked %s → goal %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(trackName), ui.Highlight.Render(goalSlug))
			return nil
		},
	}
}

func newTrackDependsOnCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "depends-on <track-name> <dependency-track>",
		Short: "Add a prerequisite track",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			trackName, dep := args[0], args[1]

			t, err := s.GetTrack(trackName)
			if err != nil {
				return err
			}
			if _, err := s.GetTrack(dep); err != nil {
				return fmt.Errorf("dependency track %q not found", dep)
			}

			for _, existing := range t.DependsOn {
				if existing == dep {
					fmt.Printf("%s already depends on %s\n", trackName, dep)
					return nil
				}
			}

			t.DependsOn = append(t.DependsOn, dep)

			allTracks, err := s.ListTracks()
			if err != nil {
				return err
			}
			for i, tt := range allTracks {
				if tt.Name == trackName {
					allTracks[i] = t
				}
			}
			if model.HasCycle(allTracks) {
				return fmt.Errorf("adding dependency %q → %q would create a cycle", trackName, dep)
			}

			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s %s now depends on %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(trackName), ui.Highlight.Render(dep))
			return nil
		},
	}
}

func newTrackMilestoneCmd(getStore func() *store.FSStore) *cobra.Command {
	ms := &cobra.Command{
		Use:   "milestone",
		Short: "Manage track milestones",
	}
	ms.AddCommand(
		newTrackMilestoneAddCmd(getStore),
		newTrackMilestoneDoneCmd(getStore),
	)
	return ms
}

func newTrackMilestoneAddCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		description string
		deadline    string
		artifact    string
	)
	cmd := &cobra.Command{
		Use:   "add <track-name>",
		Short: "Add a milestone to a track",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			t, err := s.GetTrack(args[0])
			if err != nil {
				return err
			}

			if description == "" {
				description = mustPrompt("Milestone description", "", func(v string) error {
					if strings.TrimSpace(v) == "" {
						return fmt.Errorf("description cannot be empty")
					}
					return nil
				})
			}

			m := model.Milestone{
				ID:                 store.Slugify(description),
				Description:        description,
				ArtifactResourceID: artifact,
			}
			if deadline != "" {
				d, err := time.Parse("2006-01-02", deadline)
				if err != nil {
					return fmt.Errorf("deadline must be YYYY-MM-DD, got %q", deadline)
				}
				m.Deadline = &d
			}

			t.Milestones = append(t.Milestones, m)
			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Added milestone to %s: %s\n",
				ui.Success.Render("✓"), ui.Highlight.Render(t.Name), m.Description)
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "milestone description")
	cmd.Flags().StringVar(&deadline, "deadline", "", "target date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}

func newTrackMilestoneDoneCmd(getStore func() *store.FSStore) *cobra.Command {
	var artifact string
	cmd := &cobra.Command{
		Use:   "done <track-name> <milestone-id>",
		Short: "Mark a track milestone complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			t, err := s.GetTrack(args[0])
			if err != nil {
				return err
			}

			found := false
			now := time.Now()
			for i, m := range t.Milestones {
				if m.ID == args[1] {
					t.Milestones[i].CompletedAt = &now
					if artifact != "" {
						t.Milestones[i].ArtifactResourceID = artifact
					}
					found = true
					break
				}
			}
			if !found {
				ids := make([]string, len(t.Milestones))
				for i, m := range t.Milestones {
					ids[i] = m.ID
				}
				return fmt.Errorf("milestone %q not found; available: %v", args[1], ids)
			}

			if err := s.SaveTrack(t); err != nil {
				return err
			}
			fmt.Printf("%s Milestone complete: %s\n", ui.Success.Render("✓"), args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}
