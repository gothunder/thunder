package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rotisserie/eris"
	"os"
	newCmd "thunder/internal/cli/new"
)

func Start() tea.Model {
	_, err := os.Stat("go.mod")
	if err != nil {
		if eris.Is(err, os.ErrNotExist) {
			return newCmd.New()
		}

		// TODO: handle this error
		panic(err)
	}

	// TODO file found
	return newCmd.New()
}
