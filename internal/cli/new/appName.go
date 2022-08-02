package new

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"thunder/internal/cli/styles"
	"thunder/internal/cli/utils"
)

func appNameModel() textinput.Model {
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Enter the application name"
	textInputModel.Focus()
	textInputModel.CharLimit = 156
	textInputModel.Width = 20

	return textInputModel
}

func appNameRender(m Model) string {
	return styles.AppStyle.Render(
		styles.TitleStyle.Render("Creating new application") + "\n" +
			m.appName.View(),
	)
}

func appNameUpdate(m Model, msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.step = stepModuleSelect
			m.appName.SetValue(utils.NormalizeToKebabOrSnakeCase(
				m.appName.Value(),
			))
			return m, nil
		}

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	m.appName, cmd = m.appName.Update(msg)
	return m, cmd
}
