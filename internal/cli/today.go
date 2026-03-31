package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

func newTodayCmd(getStore func() *store.FSStore) *cobra.Command {
	return &cobra.Command{
		Use:   "today",
		Short: "Daily briefing — streak, yesterday's activity, and today's suggestions",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := getStore()
			now := time.Now()
			today := truncateDay(now)
			yesterday := today.AddDate(0, 0, -1)

			// --- Header: date + streak ---
			streak, err := s.Streak(now)
			if err != nil {
				return err
			}
			fmt.Println(ui.Bold.Render(now.Format("Monday, 2 January 2006")))
			if streak > 0 {
				fmt.Printf("%s\n", ui.Highlight.Render(fmt.Sprintf("%d-day streak", streak)))
			} else {
				fmt.Printf("%s\n", ui.Muted.Render("No active streak — start one today"))
			}
			fmt.Println()

			// --- Yesterday ---
			ySessions, err := s.SessionsOnDate(yesterday)
			if err != nil {
				return err
			}
			if len(ySessions) > 0 {
				fmt.Println(ui.Bold.Render("Yesterday"))
				printSessionSummary(s, ySessions)
				fmt.Println()
			}

			// --- Today so far ---
			todaySessions, err := s.SessionsOnDate(today)
			if err != nil {
				return err
			}
			if len(todaySessions) > 0 {
				fmt.Println(ui.Bold.Render("Today so far"))
				printSessionSummary(s, todaySessions)
				fmt.Println()
			}

			// --- Suggestions ---
			all, err := s.ListResources(store.ResourceFilter{})
			if err != nil {
				return err
			}

			var candidates []model.Resource
			for _, r := range all {
				if r.Status != model.StatusDone {
					candidates = append(candidates, r)
				}
			}

			if len(candidates) == 0 {
				fmt.Println(ui.Success.Render("All resources complete — time to add more!"))
				return nil
			}

			suggestions := rankSuggestions(candidates, 5)

			fmt.Println(ui.Bold.Render("Suggested for today"))
			for _, r := range suggestions {
				est := ""
				if r.EstimatedMinutes > 0 {
					est = "  " + ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes))
				}
				trackPart := ""
				if r.Track != "" && r.Track != "inbox" {
					trackPart = "  " + ui.Dim.Render("["+r.Track+"]")
				}
				fmt.Printf("  %s  %s  %s%s%s\n",
					ui.StatusBadge(string(r.Status)),
					ui.TypeBadge(string(r.Type)),
					r.Title,
					est,
					trackPart,
				)
			}
			fmt.Println()
			fmt.Println(ui.Muted.Render("Run  qlog focus  to start a session"))

			return nil
		},
	}
}

// printSessionSummary prints a compact summary of a set of sessions.
func printSessionSummary(s *store.FSStore, sessions []model.Session) {
	totalMins := 0
	for _, sess := range sessions {
		totalMins += sess.ActualSecs / 60
	}

	sessionWord := "session"
	if len(sessions) > 1 {
		sessionWord = "sessions"
	}
	fmt.Printf("  %s  ·  %d min\n", ui.Dim.Render(fmt.Sprintf("%d %s", len(sessions), sessionWord)), totalMins)

	// Collect unique resource IDs touched (opened or status-changed)
	seen := map[string]bool{}
	var touchedIDs []string
	for _, sess := range sessions {
		for _, id := range sess.Opened {
			if !seen[id] {
				seen[id] = true
				touchedIDs = append(touchedIDs, id)
			}
		}
		for _, sc := range sess.StatusChanges {
			if !seen[sc.ResourceID] {
				seen[sc.ResourceID] = true
				touchedIDs = append(touchedIDs, sc.ResourceID)
			}
		}
	}

	if len(touchedIDs) == 0 {
		return
	}

	// Build a map of status changes for display
	finalStatus := map[string]model.Status{}
	for _, sess := range sessions {
		for _, sc := range sess.StatusChanges {
			finalStatus[sc.ResourceID] = sc.To
		}
	}

	openedSet := map[string]bool{}
	for _, sess := range sessions {
		for _, id := range sess.Opened {
			openedSet[id] = true
		}
	}

	for _, id := range touchedIDs {
		r, err := s.GetResource(id)
		title := id
		rType := ""
		if err == nil {
			title = r.Title
			rType = "  " + ui.TypeBadge(string(r.Type))
		}

		parts := []string{}
		if openedSet[id] {
			parts = append(parts, "opened")
		}
		if to, ok := finalStatus[id]; ok {
			parts = append(parts, fmt.Sprintf("→ %s", to))
		}

		annotation := ""
		if len(parts) > 0 {
			annotation = "  " + ui.Muted.Render(strings.Join(parts, " · "))
		}

		fmt.Printf("    %s%s%s\n", title, rType, annotation)
	}
}

// rankSuggestions orders candidates for the daily suggestion list.
// In-progress items first, then by priority (1=highest, 0=unset sorts last), then by estimated time.
func rankSuggestions(candidates []model.Resource, n int) []model.Resource {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]

		// In-progress first
		aIP := a.Status == model.StatusInProgress
		bIP := b.Status == model.StatusInProgress
		if aIP != bIP {
			return aIP
		}

		// Priority (1 highest, 0 = unset → sort last)
		ap := a.Priority
		bp := b.Priority
		if ap == 0 {
			ap = 999
		}
		if bp == 0 {
			bp = 999
		}
		if ap != bp {
			return ap < bp
		}

		// Shorter first as tiebreak
		return a.EstimatedMinutes < b.EstimatedMinutes
	})

	if len(candidates) > n {
		return candidates[:n]
	}
	return candidates
}

func truncateDay(t time.Time) time.Time {
	y, m, d := t.Local().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
