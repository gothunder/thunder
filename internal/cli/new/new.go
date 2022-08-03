package new

import (
	tea "github.com/charmbracelet/bubbletea"
	"thunder/internal/cli/styles"
	"time"
)

type Model struct {
	Ticks     int
	AppName   string
	Generated bool
}

func New() tea.Model {
	return Model{}
}

type tickMsg struct{}
type viewChangedMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
func viewChanged() tea.Msg {
	return viewChangedMsg{}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Generated = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if m.AppName != "" {
		return updateCreateApp(msg, m)
	}
	return updateAppName(msg, m)
}

func (m Model) View() string {
	var s string
	if m.Generated {
		return styles.AppStyle.Render(
			"App created successfully!",
		)
	}

	if m.AppName != "" {
		s = createAppView(m)
	} else {
		s = appNameView(m)
	}

	return styles.AppStyle.Render(
		s,
	)
}
