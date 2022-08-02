package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"thunder/internal/cli"
)

func main() {
	if err := tea.NewProgram(cli.Start()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
