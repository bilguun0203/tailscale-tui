package nodedetails

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bilguun0203/tailscale-tui/internal/ts"
	actionlist "github.com/bilguun0203/tailscale-tui/internal/tui/action_list"
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
	tailStatus  *ipnstate.Status
	nodeID      tsKey.NodePublic
	keyMap      keymap.KeyMap
	w, h        int
	help        help.Model
	actionsList actionlist.Model
	helpH       int
	detailH     int
	contentH    int
	messages    []string
	pingCount   int
}

type BackMsg bool

func (m Model) getCurrentNode() *ipnstate.PeerStatus {
	node, ok := m.tailStatus.Peer[m.nodeID]
	if !ok && m.tailStatus.Self.PublicKey == m.nodeID {
		node = m.tailStatus.Self
		ok = true
	}
	if ok {
		return node
	}
	return nil
}

func (m *Model) updateKeybindings() {
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
	node := m.getCurrentNode()
	if node != nil {
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
				cmd = types.NewStatusMsg(status)
				cmds = append(cmds, cmd)
			}
		}
	}
	switch {
	case key.Matches(msg, m.keyMap.Enter):
		if m.actionsList.SelectedItem().Value() == ts.ConnectAction {
			cmd = func() tea.Msg {
				return ts.ToggleConnectionMsg(true)
			}
		} else if m.actionsList.SelectedItem().Value() == ts.PingAction {
			node := m.getCurrentNode()
			if node != nil {
				cmd = func() tea.Msg {
					return ts.PingMsg(node.TailscaleIPs[0])
				}
			}
		}
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

func (m Model) messagesView() string {
	v := lipgloss.JoinVertical(
		lipgloss.Left,
		constants.PrimaryTitleStyle.Render("Messages"),
		constants.NormalTextStyle.Margin(1).Render(strings.Join(m.messages, "\n")))
	return lipgloss.NewStyle().Width(m.w / 2).Height(m.contentH).Render(v)
}

func (m *Model) SetSize(w int, h int) {
	m.w = w
	m.h = h
	m.helpH = lipgloss.Height(m.help.View(m.keyMap))
	m.detailH = lipgloss.Height(NodeDetailRender(m.tailStatus, m.nodeID, ""))
	m.contentH = m.h - m.helpH - m.detailH
	m.actionsList.SetSize(m.w/2, m.contentH)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case ts.PingMsg:
		if m.pingCount <= 0 {
			m.pingCount = 10
			m.messages = []string{}
			m.messages = append(m.messages, fmt.Sprintf("> Pinging %s. (max: 10 or until direct)", netip.Addr(msg)))
			cmds = append(cmds, types.NewStatusMsg("Pinging..."))
		}
		m.pingCount -= 1
		pr, err := ts.Ping(netip.Addr(msg))
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				m.messages = append(m.messages, fmt.Sprintf("ping %q timed out", netip.Addr(msg)))
			} else {
				m.messages = append(m.messages, fmt.Sprintf("error: %s", err))
			}
		}
		if pr != nil {
			prmsg, err := ts.PingResultString(pr)
			if err != nil {
				m.messages = append(m.messages, err.Error())
			} else {
				m.messages = append(m.messages, constants.DimmedTextStyle.Render(prmsg))
			}
			if pr.Endpoint != "" {
				m.pingCount = -1
			}
			if m.pingCount > 0 {
				cmds = append(cmds, func() tea.Msg { return msg })
			} else if m.pingCount == -1 {
				m.messages = append(m.messages, "\nDone!")
				cmds = append(cmds, types.NewStatusMsg("Pinging finished."))
			} else {
				m.messages = append(m.messages, "\nDone, direct connection not established!")
				cmds = append(cmds, types.NewStatusMsg("Pinging finished, direct connectsion not established."))
			}
		}
	case tea.KeyMsg:
		var kcmds []tea.Cmd
		m, kcmds = m.keyBindingsHandler(msg)
		cmds = append(cmds, kcmds...)
		m.updateKeybindings()
	default:
	}

	maxMessageCount := m.contentH - 3
	messageCount := len(m.messages)
	if messageCount > maxMessageCount {
		beg := messageCount - maxMessageCount
		m.messages = m.messages[beg:]
	}
	m.actionsList, cmd = m.actionsList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		NodeDetailRender(m.tailStatus, m.nodeID, ""),
		lipgloss.JoinHorizontal(lipgloss.Top, m.actionsList.View(), m.messagesView()),
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

	var actionItems []actionlist.ActionListItem
	if m.tailStatus != nil {
		if m.tailStatus.Self.PublicKey == m.nodeID {
			connection := m.tailStatus.Self.Online
			// offerExitNode := m.tailStatus.Self.ExitNodeOption
			actionItems = []actionlist.ActionListItem{
				actionlist.NewActionListItem("> Tailscale", fmt.Sprintf("Connection: %t", connection), ts.ConnectAction),
				// actionlist.NewActionListItem("> Offer Exit Node", fmt.Sprintf("%t", offerExitNode), ts.OfferExitNode),
			}
		} else {
			actionItems = []actionlist.ActionListItem{
				actionlist.NewActionListItem("> Ping", "run tailscale ping", ts.PingAction),
			}
		}
	}

	m.updateKeybindings()
	m.actionsList = actionlist.New(actionItems, m.w/2, m.h)
	m.SetSize(m.w, m.h)
	return m
}
