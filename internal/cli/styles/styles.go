package styles

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2, 2, 2)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1).
			MarginBottom(1)

	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8f8f2")).
				Background(lipgloss.Color("#ff5555"))
)
