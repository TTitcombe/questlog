package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newAddCmd(getStore func() *store.FSStore) *cobra.Command {
	var (
		quick    string
		title    string
		rtype    string
		url      string
		tags     string
		track    string
		minutes  int
		priority int
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a learning resource or idea",
		Example: `  qlog add --quick "interesting idea about transformers"
  qlog add --title "Attention is All You Need" --type paper --track llm --minutes 45`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()

			// Quick capture: straight to inbox
			if quick != "" {
				r := model.Resource{
					Title:  quick,
					Type:   model.TypeIdea,
					Track:  "inbox",
					Added:  time.Now(),
					Status: model.StatusUnread,
				}
				if err := s.SaveResource(r); err != nil {
					return err
				}
				fmt.Printf("%s Captured to inbox: %s\n", ui.Success.Render("✓"), quick)
				return nil
			}

			// If flags supplied, use them; otherwise prompt interactively.
			if title == "" {
				title = mustPrompt("Title", "", func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title cannot be empty")
					}
					return nil
				})
			}

			if rtype == "" {
				rtype = mustSelect("Type", []string{"paper", "video", "book", "article", "note", "idea"})
			}

			if url == "" {
				url = optionalPrompt("URL (optional)")
			}

			if tags == "" {
				tags = optionalPrompt("Tags (comma-separated, optional)")
			}

			if !cmd.Flags().Changed("track") {
				tracks, _ := s.ListTracks()
				choices := []string{"inbox"}
				for _, t := range tracks {
					choices = append(choices, t.Name)
				}
				track = mustSelect("Track (inbox = uncategorised)", choices)
			}

			rt := model.ResourceType(rtype)
			if minutes == 0 && !cmd.Flags().Changed("minutes") {
				def := rt.DefaultMinutes()
				raw := optionalPromptDefault("Estimated minutes", strconv.Itoa(def))
				if raw != "" {
					minutes, _ = strconv.Atoi(raw)
				}
			}

			var tagList []string
			for _, t := range strings.Split(tags, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tagList = append(tagList, t)
				}
			}

			r := model.Resource{
				Title:            title,
				Type:             rt,
				URL:              url,
				Tags:             tagList,
				Track:            track,
				Added:            time.Now(),
				EstimatedMinutes: minutes,
				Status:           model.StatusUnread,
				Priority:         priority,
			}
			if err := s.SaveResource(r); err != nil {
				return err
			}
			fmt.Printf("%s Added %s to %s\n", ui.Success.Render("✓"), ui.Bold.Render(title), ui.Highlight.Render(track))
			return nil
		},
	}

	cmd.Flags().StringVarP(&quick, "quick", "q", "", "quick capture to inbox (no prompts)")
	cmd.Flags().StringVar(&title, "title", "", "resource title")
	cmd.Flags().StringVar(&rtype, "type", "", "resource type (paper|video|book|article|note|idea)")
	cmd.Flags().StringVar(&url, "url", "", "URL")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")
	cmd.Flags().StringVarP(&track, "track", "t", "inbox", "track name (default: inbox)")
	cmd.Flags().IntVar(&minutes, "minutes", 0, "estimated minutes to complete")
	cmd.Flags().IntVar(&priority, "priority", 0, "priority 1 (highest) to 5 (lowest), 0 = unset")

	return cmd
}

func mustPrompt(label, defaultVal string, validate func(string) error) string {
	p := promptui.Prompt{
		Label:    label,
		Default:  defaultVal,
		Validate: validate,
	}
	result, err := p.Run()
	if err != nil {
		fmt.Println("\nAborted.")
		panic(err)
	}
	return result
}

func optionalPrompt(label string) string {
	p := promptui.Prompt{Label: label}
	result, _ := p.Run()
	return strings.TrimSpace(result)
}

func optionalPromptDefault(label, defaultVal string) string {
	p := promptui.Prompt{Label: label, Default: defaultVal}
	result, _ := p.Run()
	return strings.TrimSpace(result)
}

func mustSelect(label string, items []string) string {
	s := promptui.Select{
		Label: label,
		Items: items,
	}
	_, result, err := s.Run()
	if err != nil {
		fmt.Println("\nAborted.")
		panic(err)
	}
	return result
}
