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

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
