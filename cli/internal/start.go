package internal

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	newCmd "github.com/gothunder/thunder/cli/internal/new"
	"github.com/rotisserie/eris"
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
