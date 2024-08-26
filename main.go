package main

import (
	"fmt"
	"os"

	"github.com/bilguun0203/tailscale-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := tui.New()
	p := tea.NewProgram(m, tea.WithAltScreen())

	fm, err := p.Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if fm.(tui.Model).Err != nil {
		fmt.Println("Error running program:", fm.(tui.Model).Err)
		os.Exit(1)
	}

	if fm.(tui.Model).ExitMessage != "" {
		fmt.Println(fm.(tui.Model).ExitMessage)
	}
}
