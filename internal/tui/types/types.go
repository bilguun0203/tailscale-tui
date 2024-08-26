package types

import tea "github.com/charmbracelet/bubbletea"

type RefreshMsg bool
type StatusMsg string
type ExitMsg string

func NewStatusMsg(msg string) func() tea.Msg {
	return func() tea.Msg { return StatusMsg(msg) }
}
