package nodedetails

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/bilguun0203/tailscale-tui/internal/tui/keymap"
	"github.com/bilguun0203/tailscale-tui/internal/tui/types"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/ipn/ipnstate"
	tsKey "tailscale.com/types/key"
)

type Model struct {
	tailStatus *ipnstate.Status
	nodeID     tsKey.NodePublic
	keyMap     keymap.KeyMap
	w, h       int
	help       help.Model
}

type BackMsg bool

func (m *Model) updateKeybindings() {
	m.keyMap.Enter.SetEnabled(false)
	m.keyMap.Refresh.SetEnabled(false)
	if m.help.ShowAll {
		m.keyMap.ShowFullHelp.SetEnabled(false)
		m.keyMap.CloseFullHelp.SetEnabled(true)
	} else {
		m.keyMap.ShowFullHelp.SetEnabled(true)
		m.keyMap.CloseFullHelp.SetEnabled(false)
	}
}

func (m Model) keyBindingsHandler(msg tea.KeyMsg) (Model, []tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	node, ok := m.tailStatus.Peer[m.nodeID]
	if !ok && m.tailStatus.Self.PublicKey == m.nodeID {
		node = m.tailStatus.Self
		ok = true
	}
	if ok {
		if key.Matches(msg, m.keyMap.CopyIpv4) || key.Matches(msg, m.keyMap.CopyIpv6) || key.Matches(msg, m.keyMap.CopyDNSName) {
			copyStr := ""
			ipCount := len(node.TailscaleIPs)
			if ipCount > 0 && key.Matches(msg, m.keyMap.CopyIpv4) {
				copyStr = node.TailscaleIPs[0].String()
			} else if ipCount > 1 && key.Matches(msg, m.keyMap.CopyIpv6) {
				copyStr = node.TailscaleIPs[1].String()
			} else if key.Matches(msg, m.keyMap.CopyDNSName) {
				copyStr = node.DNSName
			}
			if copyStr != "" {
				clipboard.WriteAll(copyStr)
				status := fmt.Sprintf("Copied \"%s\"!", constants.PrimaryTextStyle.Underline(true).Render(copyStr))
				cmd = func() tea.Msg { return types.StatusMsg(status) }
				cmds = append(cmds, cmd)
			}
		}
	}
	switch {
	case key.Matches(msg, m.keyMap.Back):
		cmd = func() tea.Msg {
			return BackMsg(true)
		}
	case key.Matches(msg, m.keyMap.Quit):
		cmd = tea.Quit
	case key.Matches(msg, m.keyMap.ShowFullHelp):
		m.help.ShowAll = true
	case key.Matches(msg, m.keyMap.CloseFullHelp):
		m.help.ShowAll = false
	}

	cmds = append(cmds, cmd)
	return m, cmds
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		var kcmds []tea.Cmd
		m, kcmds = m.keyBindingsHandler(msg)
		cmds = append(cmds, kcmds...)
		m.updateKeybindings()
	default:
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	helpHeight := lipgloss.Height(m.help.View(m.keyMap))
	detailHeight := m.h - helpHeight
	return lipgloss.JoinVertical(
		lipgloss.Left, lipgloss.NewStyle().Height(detailHeight).Render(NodeDetailRender(m.tailStatus, m.nodeID, "")),
		lipgloss.NewStyle().Margin(0, 2).Render(m.help.View(m.keyMap)),
	)

}

func New(status *ipnstate.Status, nodeID tsKey.NodePublic, w, h int) Model {
	m := Model{
		keyMap:     keymap.NewKeyMap(),
		tailStatus: status,
		nodeID:     nodeID,
		w:          w,
		h:          h,
		help:       help.New(),
	}
	m.updateKeybindings()
	return m
}
