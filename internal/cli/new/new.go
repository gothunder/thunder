package new

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"thunder/internal/cli/styles"
	"time"
)

type steps int

const (
	stepAppName steps = iota
	stepModuleSelect
)

type Model struct {
	appName textinput.Model
	step    steps
	err     error
}

func New() tea.Model {
	return Model{
		appName: appNameModel(),
		step:    stepAppName,
		err:     nil,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.step {
	case stepAppName:
		return appNameUpdate(m, msg)
	}

	time.Sleep(time.Second * 3)

	return m, tea.Quit
}

func (m Model) View() string {
	if m.err != nil {
		return styles.StatusMessageStyle.Render(m.err.Error())
	}

	switch m.step {
	case stepAppName:
		return appNameRender(m)
	}

	return m.appName.Value()
}
