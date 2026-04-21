package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newGoalCmd(getStore func() *store.FSStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "goal",
		Short: "Manage learning goals",
	}
	cmd.AddCommand(
		newGoalNewCmd(getStore),
		newGoalListCmd(getStore),
		newGoalShowCmd(getStore),
		newGoalMilestoneCmd(getStore),
	)
	return cmd
}

func newGoalNewCmd(getStore func() *store.FSStore) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new learning goal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			title := args[0]
			slug := store.Slugify(title)

			if _, err := s.LoadGoal(slug); err == nil {
				return fmt.Errorf("goal %q already exists", slug)
			}

			g := model.Goal{
				Slug:        slug,
				Title:       title,
				Description: description,
				Created:     time.Now(),
			}
			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Created goal: %s\n", ui.Success.Render("✓"), ui.Highlight.Render(title))
			fmt.Printf("  Slug: %s\n", ui.Dim.Render(slug))
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "goal description")
	return cmd
}

func newGoalListCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all goals",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			goals, err := s.ListGoals()
			if err != nil {
				return err
			}
			if len(goals) == 0 {
				fmt.Println(ui.Muted.Render("No goals yet. Create one with: qlog goal new \"<title>\""))
				return nil
			}
			tracks, _ := s.ListTracks()
			for _, g := range goals {
				linked := 0
				for _, t := range tracks {
					if t.GoalSlug == g.Slug {
						linked++
					}
				}
				fmt.Printf("  %s  %s\n", ui.Highlight.Render(g.Title), ui.Dim.Render(g.Slug))
				if g.Description != "" {
					fmt.Printf("    %s\n", ui.Muted.Render(g.Description))
				}
				pending := 0
				for _, m := range g.Milestones {
					if m.CompletedAt == nil {
						pending++
					}
				}
				fmt.Printf("    %d tracks · %d milestones (%d pending)\n\n",
					linked, len(g.Milestones), pending)
			}
			return nil
		},
	}
}

func newGoalShowCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "show <slug>",
		Short: "Show goal details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
			}

			fmt.Printf("%s\n", ui.Bold.Render(g.Title))
			if g.Description != "" {
				fmt.Printf("%s\n", ui.Muted.Render(g.Description))
			}
			fmt.Printf("Slug: %s  Created: %s\n\n", ui.Dim.Render(g.Slug), g.Created.Format("2006-01-02"))

			if len(g.Milestones) > 0 {
				fmt.Println(ui.Bold.Render("Goal milestones:"))
				for _, m := range g.Milestones {
					check := "○"
					if m.CompletedAt != nil {
						check = ui.Success.Render("✓")
					}
					deadline := ""
					if m.Deadline != nil {
						deadline = ui.Warning.Render(fmt.Sprintf("  due %s", m.Deadline.Format("2006-01-02")))
					}
					fmt.Printf("  %s %s%s\n", check, m.Description, deadline)
					fmt.Printf("    %s\n", ui.Dim.Render(m.ID))
				}
				fmt.Println()
			}

			tracks, _ := s.ListTracks()
			fmt.Println(ui.Bold.Render("Tracks:"))
			found := false
			for _, t := range tracks {
				if t.GoalSlug != g.Slug {
					continue
				}
				found = true
				deps := ""
				if len(t.DependsOn) > 0 {
					deps = ui.Dim.Render(fmt.Sprintf("  needs: %s", strings.Join(t.DependsOn, ", ")))
				}
				fmt.Printf("  %s%s\n", ui.Highlight.Render(t.Name), deps)
			}
			if !found {
				fmt.Println(ui.Muted.Render("  No tracks linked yet. Use: qlog track set-goal <name> " + g.Slug))
			}
			return nil
		},
	}
}

func newGoalMilestoneCmd(getStore func() *store.FSStore) *cobra.Command {
	ms := &cobra.Command{
		Use:   "milestone",
		Short: "Manage goal milestones",
	}
	ms.AddCommand(
		newGoalMilestoneAddCmd(getStore),
		newGoalMilestoneDoneCmd(getStore),
	)
	return ms
}

func newGoalMilestoneAddCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		description string
		deadline    string
		artifact    string
	)
	cmd := &cobra.Command{
		Use:   "add <goal-slug>",
		Short: "Add a milestone to a goal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
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

			g.Milestones = append(g.Milestones, m)
			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Added milestone: %s\n", ui.Success.Render("✓"), m.Description)
			return nil
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "milestone description")
	cmd.Flags().StringVar(&deadline, "deadline", "", "target date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID that proves this milestone")
	return cmd
}

func newGoalMilestoneDoneCmd(getStore func() *store.FSStore) *cobra.Command {
	var artifact string
	cmd := &cobra.Command{
		Use:   "done <goal-slug> <milestone-id>",
		Short: "Mark a goal milestone complete",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			g, err := s.LoadGoal(args[0])
			if err != nil {
				return fmt.Errorf("goal %q not found", args[0])
			}

			found := false
			now := time.Now()
			for i, m := range g.Milestones {
				if m.ID == args[1] {
					g.Milestones[i].CompletedAt = &now
					if artifact != "" {
						g.Milestones[i].ArtifactResourceID = artifact
					}
					found = true
					break
				}
			}
			if !found {
				ids := make([]string, len(g.Milestones))
				for i, m := range g.Milestones {
					ids[i] = m.ID
				}
				data, _ := json.Marshal(ids)
				return fmt.Errorf("milestone %q not found; available: %s", args[1], data)
			}

			if err := s.SaveGoal(g); err != nil {
				return err
			}
			fmt.Printf("%s Milestone complete: %s\n", ui.Success.Render("✓"), args[1])
			return nil
		},
	}
	cmd.Flags().StringVar(&artifact, "artifact", "", "questlog resource ID proving completion")
	return cmd
}
