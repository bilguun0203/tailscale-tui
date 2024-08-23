package nodelist

import (
	"fmt"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bilguun0203/tailscale-tui/internal/ts"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/bilguun0203/tailscale-tui/internal/tui/keymap"
	"github.com/bilguun0203/tailscale-tui/internal/tui/types"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/ipn/ipnstate"
	tsKey "tailscale.com/types/key"
)

type listItem struct {
	title, desc string
	status      *ipnstate.PeerStatus
}

func (i listItem) Title() string                { return i.title }
func (i listItem) Description() string          { return i.desc }
func (i listItem) Status() *ipnstate.PeerStatus { return i.status }
func (i listItem) FilterValue() string          { return i.title + " " + i.desc }

type Model struct {
	tailStatus *ipnstate.Status
	exitNode   string
	list       list.Model
	keyMap     keymap.KeyMap
	w          int
	h          int
}

func (m *Model) SetSize(w int, h int) {
	m.w = w
	m.h = h
	m.list.SetSize(w, h)
}

type NodeSelectedMsg tsKey.NodePublic

func (m *Model) updateKeybindings() {
	if m.list.SelectedItem() != nil {
		m.keyMap.CopyIpv4.SetEnabled(true)
		m.keyMap.CopyIpv6.SetEnabled(true)
		m.keyMap.CopyDNSName.SetEnabled(true)
		m.keyMap.Enter.SetEnabled(true)
	} else {
		m.keyMap.CopyIpv4.SetEnabled(false)
		m.keyMap.CopyIpv6.SetEnabled(false)
		m.keyMap.CopyDNSName.SetEnabled(false)
		m.keyMap.Enter.SetEnabled(false)
	}
	m.keyMap.Back.SetEnabled(false)
	m.keyMap.Quit.SetEnabled(false)
	m.keyMap.ShowFullHelp.SetEnabled(false)
	m.keyMap.CloseFullHelp.SetEnabled(false)
	m.keyMap.ForceQuit.SetEnabled(false)
	m.list.KeyMap.NextPage.SetEnabled(false)
	m.list.KeyMap.PrevPage.SetEnabled(false)
}

func (m Model) keyBindingsHandler(msg tea.KeyMsg) (Model, []tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	if key.Matches(msg, m.keyMap.CopyIpv4) || key.Matches(msg, m.keyMap.CopyIpv6) || key.Matches(msg, m.keyMap.CopyDNSName) {
		copyStr := ""
		ipCount := len(m.list.SelectedItem().(listItem).status.TailscaleIPs)
		if ipCount > 0 && key.Matches(msg, m.keyMap.CopyIpv4) {
			copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[0].String()
		} else if ipCount > 1 && key.Matches(msg, m.keyMap.CopyIpv6) {
			copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[1].String()
		} else if key.Matches(msg, m.keyMap.CopyDNSName) {
			copyStr = m.list.SelectedItem().(listItem).status.DNSName
		}
		if copyStr == "" {
			m.list.NewStatusMessage("Sorry, nothing to copy.")
			cmd = func() tea.Msg { return types.StatusMsg("Sorry, nothing to copy.") }
			cmds = append(cmds, cmd)
		} else {
			err := clipboard.WriteAll(copyStr)
			status := fmt.Sprintf("Copied \"%s\"!", constants.PrimaryTextStyle.Underline(true).Render(copyStr))
			if err != nil {
				status = fmt.Sprintf("Sorry, error occured: %s", err)
			}
			m.list.NewStatusMessage(status)
			cmd = func() tea.Msg { return types.StatusMsg(status) }
			cmds = append(cmds, cmd)
		}
	}
	if key.Matches(msg, m.keyMap.Refresh) {
		cmd = func() tea.Msg { return types.RefreshMsg(true) }
		cmds = append(cmds, cmd)
	}
	if key.Matches(msg, m.keyMap.Enter) {
		cmd = func() tea.Msg { return NodeSelectedMsg(m.list.SelectedItem().(listItem).status.PublicKey) }
		cmds = append(cmds, cmd)
	}
	if key.Matches(msg, m.keyMap.TSUp) {
		ts.SetTSStatus(true)
		cmd = func() tea.Msg { return types.RefreshMsg(true) }
		cmds = append(cmds, cmd)
	}
	if key.Matches(msg, m.keyMap.TSDown) {
		ts.SetTSStatus(false)
		cmd = func() tea.Msg { return types.RefreshMsg(true) }
		cmds = append(cmds, cmd)
	}
	return m, cmds
}


func (m *Model) getItems() []list.Item {
	items := []list.Item{}
	peers := []*ipnstate.PeerStatus{}

	if m.tailStatus == nil {
		return items
	}

	for _, v := range m.tailStatus.Peer {
		peers = append(peers, v)
	}
	sort.Slice(peers, func(i, j int) bool {
		ownNode1 := m.tailStatus.Self.UserID == peers[i].UserID
		ownNode2 := m.tailStatus.Self.UserID == peers[j].UserID
		first := fmt.Sprint(!ownNode1) + strings.ToLower(peers[i].HostName)
		second := fmt.Sprint(!ownNode2) + strings.ToLower(peers[j].HostName)
		return first < second
	})

	peers = append([]*ipnstate.PeerStatus{m.tailStatus.Self}, peers...)

	m.exitNode = ""
	for _, v := range peers {
		state := constants.DangerTextStyle.Render("●")
		if v.Online {
			state = constants.SuccessTextStyle.Render("●")
		}
		hostName := v.HostName
		owner := "my device"
		if v.ID == m.tailStatus.Self.ID {
			hostName = "◆ " + hostName
			owner = "this device"
		}
		if v.UserID != m.tailStatus.Self.UserID {
			owner = "from:" + m.tailStatus.User[v.UserID].LoginName
		}
		owner = constants.DimmedTextStyle.Render("[" + owner + "]")
		exitNode := ""
		if v.ExitNodeOption {
			exitNode = constants.DimmedTextStyle.Bold(true).Render("[→]")
		}
		if v.ExitNode {
			m.exitNode = v.PublicKey.String()
			exitNode = constants.SuccessTextStyle.Bold(true).Render("[→]")
		}
		os := constants.NormalTextStyle.Render(v.OS)
		title := fmt.Sprintf("%s %s %s %s %s", hostName, state, owner, os, exitNode)
		desc := "- "
		var ips []string
		for _, ip := range v.TailscaleIPs {
			ips = append(ips, ip.String())
		}
		ips = append(ips, v.DNSName)
		desc += strings.Join(ips, " | ")
		items = append(items, listItem{title: title, desc: desc, status: v})
	}
	return items
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case ts.StatusDataMsg:
		m.tailStatus = msg
		if m.tailStatus != nil {
			cmds = append(cmds, m.list.SetItems(m.getItems()))
		}
		m.list.StopSpinner()
	case tea.KeyMsg:
		var kcmds []tea.Cmd
		m, kcmds = m.keyBindingsHandler(msg)
		cmds = append(cmds, kcmds...)
	default:
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	m.updateKeybindings()
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.list.View()
}

func New(status *ipnstate.Status, w, h int) Model {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = d.Styles.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(constants.PrimaryTextStyle.GetForeground()).
		Foreground(constants.PrimaryTextStyle.GetForeground()).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle
	d.Styles.DimmedTitle = constants.DimmedTextStyle.Padding(0, 0, 0, 2)
	d.Styles.DimmedDesc = d.Styles.DimmedTitle.Foreground(constants.MutedTextStyle.GetForeground())
	d.SetHeight(2)
	d.SetSpacing(1)
	m := Model{
		list:       list.New([]list.Item{}, d, w, h),
		keyMap:     keymap.NewKeyMap(),
		tailStatus: status,
		w:          w,
		h:          h,
	}
	m.list.SetSpinner(spinner.Dot)
	m.list.StartSpinner()
	m.list.SetHeight(h)

	m.list.SetItems(m.getItems())

	m.list.Title = "Nodes"
	m.list.Styles.Title = constants.PrimaryTitleStyle
	m.list.FilterInput.PromptStyle = constants.PrimaryTextStyle
	m.list.FilterInput.Cursor.Style = constants.PrimaryTextStyle
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keyMap.CopyIpv4,
			m.keyMap.Refresh,
			m.keyMap.Enter,
		}
	}
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keyMap.CopyIpv4,
			m.keyMap.CopyIpv6,
			m.keyMap.CopyDNSName,
			m.keyMap.Refresh,
			m.keyMap.Enter,
		}
	}
	m.updateKeybindings()
	return m
}
