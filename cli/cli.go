package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gothunder/thunder/cli/internal"
)

func main() {
	if err := tea.NewProgram(internal.Start()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
