package ui

import "github.com/charmbracelet/lipgloss"

var (
	Bold      = lipgloss.NewStyle().Bold(true)
	Dim       = lipgloss.NewStyle().Faint(true)
	Highlight = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	Success   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	Warning   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Muted     = lipgloss.NewStyle().Faint(true)
	Error     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	TypeColors = map[string]lipgloss.Style{
		"paper":   lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"video":   lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
		"book":    lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"article": lipgloss.NewStyle().Foreground(lipgloss.Color("45")),
		"note":    lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		"idea":    lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
	}

	StatusColors = map[string]lipgloss.Style{
		"unread":      lipgloss.NewStyle().Foreground(lipgloss.Color("248")),
		"in-progress": lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"done":        lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
	}
)

func TypeBadge(t string) string {
	style, ok := TypeColors[t]
	if !ok {
		style = Dim
	}
	return style.Render("[" + t + "]")
}

func StatusBadge(s string) string {
	style, ok := StatusColors[s]
	if !ok {
		style = Dim
	}
	switch s {
	case "done":
		return style.Render("✓ done")
	case "in-progress":
		return style.Render("… in-progress")
	default:
		return style.Render("○ unread")
	}
}

func ProgressBar(pct int, width int) string {
	if width <= 0 {
		width = 20
	}
	filled := pct * width / 100
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
