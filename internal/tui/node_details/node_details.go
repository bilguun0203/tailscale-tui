package nodedetails

import (
	"fmt"
	"strings"

	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/bilguun0203/tailscale-tui/internal/tui/keymap"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/ipn/ipnstate"
	tskey "tailscale.com/types/key"
)

type Model struct {
	tailStatus *ipnstate.Status
	nodeID     tskey.NodePublic
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

func (m Model) detailView() string {
	title := constants.AltTitleStyle.Render(" Node info ")
	hostname := constants.AltAccentTextStyle.Render("Host: ")
	userInfo := "??? <???>"
	ips := constants.AltAccentTextStyle.Render("IPs: ")
	offersExitNode := "no"
	exitNode := constants.AltAccentTextStyle.Render("Exit node: ")
	asExitNode := ""
	if m.tailStatus != nil {
		node, ok := m.tailStatus.Peer[m.nodeID]
		if !ok && m.tailStatus.Self.PublicKey == m.nodeID {
			node = m.tailStatus.Self
			ok = true
		}
		if ok {
			if user, ok := m.tailStatus.User[node.UserID]; ok {
				userInfo = fmt.Sprintf("%s <%s>", user.DisplayName, user.LoginName)
			}
			if node.ExitNodeOption {
				offersExitNode = constants.WarningTextStyle.Render("yes")
			}
			ipList := []string{}
			for _, ip := range node.TailscaleIPs {
				ipList = append(ipList, ip.String())
			}
			ipList = append(ipList, node.DNSName)
			ips += strings.Join(ipList, " | ")
			hostname += node.HostName + " (" + node.OS + ")"
			exitNode += constants.MutedTextStyle.Render("offers:") + offersExitNode
			if node.ExitNode {
				asExitNode = constants.WarningTextStyle.Render("~ This node is currently being used as an exit node.")
			}
		}
	}
	body := constants.NormalTextStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s\n%s", userInfo, hostname, ips, exitNode, asExitNode))
	return constants.HeaderStyle.Render(fmt.Sprintf("%s\n\n%s", title, body))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// var cmd tea.Cmd
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
		lipgloss.Left, lipgloss.NewStyle().Height(detailHeight).Render(m.detailView()),
		lipgloss.NewStyle().Margin(0, 2).Render(m.help.View(m.keyMap)),
	)

}

func New(status *ipnstate.Status, nodeID tskey.NodePublic, w, h int) Model {
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
