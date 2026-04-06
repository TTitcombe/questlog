package cli

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/TTitcombe/questlog/internal/cli/ui"
	"github.com/TTitcombe/questlog/internal/model"
	"github.com/TTitcombe/questlog/internal/store"
)

type focusState int

const (
	focusStateNormal focusState = iota
	focusStateStatusPick
	focusStateTimesUp
	focusStateNoteInput
	focusStateRatingPick
	focusStateBrowse
)

type focusTickMsg time.Time

type focusModel struct {
	session   []model.Resource // resources that fit in the session
	noEst     []model.Resource // resources with no time estimate
	cursor    int
	remaining time.Duration
	state     focusState
	store     *store.FSStore
	// status picker
	statusCursor int
	// rating picker
	ratingCursor int
	ratingPrompt string // header shown above rating picker
	// note input
	noteInput string
	// browse picker
	browseList   []model.Resource
	browseCursor int
	// transient feedback line
	notice string
	// session tracking
	startedAt     time.Time
	opened        []string             // resource IDs opened in browser (deduplicated)
	statusChanges []model.StatusChange // status changes made during session
}

var focusStatuses = []model.Status{model.StatusUnread, model.StatusInProgress, model.StatusDone}

var focusRatings = []struct {
	value int
	label string
}{
	{1, "+1  valuable"},
	{0, " 0  neutral"},
	{-1, "-1  not worth it"},
}

func newFocusModel(session, noEst, all []model.Resource, minutes int, s *store.FSStore) focusModel {
	return focusModel{
		session:    session,
		noEst:      noEst,
		browseList: all,
		remaining:  time.Duration(minutes) * time.Minute,
		store:      s,
		startedAt:  time.Now(),
	}
}

func (m focusModel) Init() tea.Cmd {
	return focusTick()
}

func focusTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return focusTickMsg(t)
	})
}

func (m focusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case focusTickMsg:
		if m.state == focusStateNormal {
			m.remaining -= time.Second
			if m.remaining <= 0 {
				m.remaining = 0
				m.state = focusStateTimesUp
			}
		}
		return m, focusTick()

	case tea.KeyMsg:
		switch m.state {
		case focusStateTimesUp:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}

		case focusStateNormal:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.notice = ""
				}
			case "down", "j":
				if m.cursor < len(m.session)-1 {
					m.cursor++
					m.notice = ""
				}
			case "enter":
				if len(m.session) == 0 {
					break
				}
				r := m.session[m.cursor]
				didOpen := false
				if r.URL != "" {
					if err := openBrowser(r.URL); err == nil {
						didOpen = true
						m.opened = appendUnique(m.opened, r.ID)
					}
				}
				// Mark in-progress if currently unread
				wasUnread := r.Status == model.StatusUnread
				if wasUnread {
					r.Status = model.StatusInProgress
					if err := m.store.SaveResource(r); err == nil {
						m.session[m.cursor] = r
						m.statusChanges = append(m.statusChanges, model.StatusChange{
							ResourceID: r.ID,
							To:         model.StatusInProgress,
						})
					}
				}
				switch {
				case didOpen && wasUnread:
					m.notice = "Opened in browser · marked in-progress"
				case didOpen:
					m.notice = "Opened in browser"
				case r.URL == "" && wasUnread:
					m.notice = "No URL — marked in-progress"
				default:
					m.notice = "Marked in-progress"
				}
			case "s":
				if len(m.session) == 0 {
					break
				}
				m.state = focusStateStatusPick
				cur := m.session[m.cursor].Status
				for i, s := range focusStatuses {
					if s == cur {
						m.statusCursor = i
						break
					}
				}
			case "b":
				if len(m.browseList) == 0 {
					break
				}
				m.browseCursor = 0
				m.state = focusStateBrowse
			case "n":
				if len(m.session) == 0 {
					break
				}
				m.noteInput = ""
				m.state = focusStateNoteInput
			case "r":
				if len(m.session) == 0 {
					break
				}
				m.ratingPrompt = fmt.Sprintf("Rate %q:", m.session[m.cursor].Title)
				m.ratingCursor = 0
				m.state = focusStateRatingPick
			}

		case focusStateStatusPick:
			switch msg.String() {
			case "up", "k":
				if m.statusCursor > 0 {
					m.statusCursor--
				}
			case "down", "j":
				if m.statusCursor < len(focusStatuses)-1 {
					m.statusCursor++
				}
			case "enter":
				newStatus := focusStatuses[m.statusCursor]
				r := m.session[m.cursor]
				r.Status = newStatus
				if err := m.store.SaveResource(r); err == nil {
					m.statusChanges = append(m.statusChanges, model.StatusChange{
						ResourceID: r.ID,
						To:         newStatus,
					})
					m.session[m.cursor] = r
					m.notice = fmt.Sprintf("Status → %s", newStatus)
				} else {
					m.notice = "Error saving status"
				}
				// Auto-prompt for rating when marking done and resource has no rating.
				if newStatus == model.StatusDone && m.session[m.cursor].Rating == nil {
					m.ratingPrompt = "How valuable was this?"
					m.ratingCursor = 0
					m.state = focusStateRatingPick
				} else {
					m.state = focusStateNormal
				}
			case "esc", "q":
				m.state = focusStateNormal
			}

		case focusStateNoteInput:
			switch msg.String() {
			case "esc":
				m.noteInput = ""
				m.state = focusStateNormal
			case "enter":
				if strings.TrimSpace(m.noteInput) != "" {
					r := m.session[m.cursor]
					r.Notes = prependNote(r.Notes, m.noteInput, time.Now())
					if err := m.store.SaveResource(r); err == nil {
						m.session[m.cursor] = r
						m.notice = "Note saved"
					} else {
						m.notice = "Error saving note"
					}
				}
				m.noteInput = ""
				m.state = focusStateNormal
			case "backspace":
				if len(m.noteInput) > 0 {
					runes := []rune(m.noteInput)
					m.noteInput = string(runes[:len(runes)-1])
				}
			default:
				if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
					m.noteInput += msg.String()
				}
			}

		case focusStateRatingPick:
			switch msg.String() {
			case "up", "k":
				if m.ratingCursor > 0 {
					m.ratingCursor--
				}
			case "down", "j":
				if m.ratingCursor < len(focusRatings)-1 {
					m.ratingCursor++
				}
			case "enter":
				v := focusRatings[m.ratingCursor].value
				r := m.session[m.cursor]
				r.Rating = &v
				if err := m.store.SaveResource(r); err == nil {
					m.session[m.cursor] = r
					m.notice = fmt.Sprintf("Rated: %s", formatRating(&v))
				} else {
					m.notice = "Error saving rating"
				}
				m.state = focusStateNormal
			case "esc", "q":
				m.state = focusStateNormal
			}

		case focusStateBrowse:
			switch msg.String() {
			case "up", "k":
				if m.browseCursor > 0 {
					m.browseCursor--
				}
			case "down", "j":
				if m.browseCursor < len(m.browseList)-1 {
					m.browseCursor++
				}
			case "enter":
				picked := m.browseList[m.browseCursor]
				// If already in session, move cursor there.
				for i, r := range m.session {
					if r.ID == picked.ID {
						m.cursor = i
						m.state = focusStateNormal
						return m, nil
					}
				}
				// Otherwise prepend to session and focus it.
				m.session = append([]model.Resource{picked}, m.session...)
				m.cursor = 0
				m.state = focusStateNormal
			case "esc", "q":
				m.state = focusStateNormal
			}
		}
	}
	return m, nil
}

func (m focusModel) View() string {
	var b strings.Builder

	// Header
	timerStr := formatFocusDuration(m.remaining)
	if m.state == focusStateTimesUp {
		timerStr = ui.Warning.Render("Time's up!")
	}
	b.WriteString(ui.Bold.Render("Focus session") + "  " + timerStr + "\n\n")

	// Resource list
	if len(m.session) == 0 {
		b.WriteString(ui.Muted.Render("Nothing fits in the session.") + "\n")
	} else {
		for i, r := range m.session {
			prefix := "  "
			if i == m.cursor && m.state != focusStateTimesUp {
				prefix = ui.Highlight.Render("▶ ")
			}

			title := r.Title
			if i == m.cursor && m.state != focusStateTimesUp {
				title = ui.Bold.Render(r.Title)
			}

			est := ""
			if r.EstimatedMinutes > 0 {
				est = "  " + ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes))
			}

			urlPart := "  " + ui.Muted.Render("no url")
			if r.URL != "" {
				urlPart = "  " + ui.Muted.Render(truncateStr(r.URL, 50))
			}

			trackPart := ""
			if r.Track != "" && r.Track != "inbox" {
				trackPart = "  " + ui.Dim.Render("["+r.Track+"]")
			}

			b.WriteString(fmt.Sprintf("%s%s  %s  %s%s%s%s\n",
				prefix,
				ui.StatusBadge(string(r.Status)),
				ui.TypeBadge(string(r.Type)),
				title,
				est,
				trackPart,
				urlPart,
			))
		}
	}

	// No-estimate extras
	if len(m.noEst) > 0 {
		b.WriteString("\n" + ui.Muted.Render("Also consider (no estimate):") + "\n")
		for _, r := range m.noEst {
			urlPart := ""
			if r.URL != "" {
				urlPart = "  " + ui.Muted.Render(truncateStr(r.URL, 50))
			}
			b.WriteString(fmt.Sprintf("  %s  %s  %s%s\n",
				ui.TypeBadge(string(r.Type)),
				ui.StatusBadge(string(r.Status)),
				r.Title,
				urlPart,
			))
		}
	}

	b.WriteString("\n")

	// Status picker
	if m.state == focusStateStatusPick {
		r := m.session[m.cursor]
		b.WriteString(ui.Bold.Render(fmt.Sprintf("Set status for %q:", r.Title)) + "\n")
		for i, s := range focusStatuses {
			cur := "  "
			if i == m.statusCursor {
				cur = ui.Highlight.Render("▶ ")
			}
			b.WriteString(cur + ui.StatusBadge(string(s)) + "\n")
		}
		b.WriteString("\n" + ui.Muted.Render("enter confirm · esc cancel") + "\n")
		return b.String()
	}

	// Rating picker
	if m.state == focusStateRatingPick {
		b.WriteString(ui.Bold.Render(m.ratingPrompt) + "\n")
		for i, opt := range focusRatings {
			cur := "  "
			if i == m.ratingCursor {
				cur = ui.Highlight.Render("▶ ")
			}
			b.WriteString(cur + opt.label + "\n")
		}
		b.WriteString("\n" + ui.Muted.Render("enter confirm · esc skip") + "\n")
		return b.String()
	}

	// Browse picker
	if m.state == focusStateBrowse {
		b.WriteString(ui.Bold.Render("Choose a resource:") + "\n\n")
		for i, r := range m.browseList {
			prefix := "  "
			if i == m.browseCursor {
				prefix = ui.Highlight.Render("▶ ")
			}
			title := r.Title
			if i == m.browseCursor {
				title = ui.Bold.Render(r.Title)
			}
			est := ""
			if r.EstimatedMinutes > 0 {
				est = "  " + ui.Dim.Render(fmt.Sprintf("~%dm", r.EstimatedMinutes))
			}
			b.WriteString(fmt.Sprintf("%s%s  %s  %s%s\n",
				prefix,
				ui.StatusBadge(string(r.Status)),
				ui.TypeBadge(string(r.Type)),
				title,
				est,
			))
		}
		b.WriteString("\n" + ui.Muted.Render("↑/↓ navigate · enter pick · esc cancel") + "\n")
		return b.String()
	}

	// Note input
	if m.state == focusStateNoteInput {
		b.WriteString(ui.Bold.Render("Add note:") + " " + m.noteInput + "▌\n")
		b.WriteString(ui.Muted.Render("enter save · esc cancel") + "\n")
		return b.String()
	}

	// Notice + footer
	if m.notice != "" {
		b.WriteString(ui.Success.Render(m.notice) + "\n")
	}
	if m.state == focusStateTimesUp {
		b.WriteString(ui.Muted.Render("Press q to exit") + "\n")
	} else if len(m.session) > 0 {
		b.WriteString(ui.Muted.Render("↑/↓ navigate · enter open+start · s status · n note · r rate · b browse · q quit") + "\n")
	} else {
		b.WriteString(ui.Muted.Render("q quit") + "\n")
	}

	return b.String()
}

func formatFocusDuration(d time.Duration) string {
	if d <= 0 {
		return "0:00"
	}
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
