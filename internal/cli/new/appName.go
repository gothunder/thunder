package new

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"thunder/internal/cli/styles"
	"thunder/internal/cli/utils"
)

var appNameInput = appNameModel()

func appNameModel() textinput.Model {
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Enter the application name"
	textInputModel.Focus()
	textInputModel.CharLimit = 156
	textInputModel.Width = 20

	return textInputModel
}

func appNameView(m Model) string {
	return styles.TitleStyle.Render("Creating new application") + "\n" +
		appNameInput.View()
}

func updateAppName(msg tea.Msg, m Model) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.AppName = utils.NormalizeToKebabOrSnakeCase(
				appNameInput.Value(),
			)
			return m, viewChanged
		}
	}

	appNameInput, cmd = appNameInput.Update(msg)
	return m, cmd
}
