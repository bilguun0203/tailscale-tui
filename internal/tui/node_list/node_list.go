package nodelist

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bilguun0203/tailscale-tui/internal/tailscale"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	"github.com/bilguun0203/tailscale-tui/internal/tui/keymap"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listItem struct {
	title, desc string
	status      tailscale.PeerStatus
}

func (i listItem) Title() string                { return i.title }
func (i listItem) Description() string          { return i.desc }
func (i listItem) Status() tailscale.PeerStatus { return i.status }
func (i listItem) FilterValue() string          { return i.title + " " + i.desc }

type Model struct {
	tailStatus *tailscale.Status
	exitNode   string
	list       list.Model
	keyMap     keymap.KeyMap
	msg        string
	w          int
	h          int
}

func (m *Model) SetSize(w int, h int) {
	m.w = w
	m.h = h
	headerHeight := lipgloss.Height(m.headerView())
	m.list.SetSize(w, h-headerHeight)
}

type NodeSelectedMsg string
type RefreshMsg bool

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
	m.keyMap.ShowFullHelp.SetEnabled(false)
	m.keyMap.CloseFullHelp.SetEnabled(false)
	m.keyMap.Back.SetEnabled(false)
	m.keyMap.Quit.SetEnabled(false)
	m.keyMap.ForceQuit.SetEnabled(false)
}

func (m Model) keyBindingsHandler(msg tea.KeyMsg) (Model, []tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	if key.Matches(msg, m.keyMap.CopyIpv4) || key.Matches(msg, m.keyMap.CopyIpv6) || key.Matches(msg, m.keyMap.CopyDNSName) {
		copyStr := ""
		ipCount := len(m.list.SelectedItem().(listItem).status.TailscaleIPs)
		if ipCount > 0 && key.Matches(msg, m.keyMap.CopyIpv4) {
			copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[0]
		} else if ipCount > 1 && key.Matches(msg, m.keyMap.CopyIpv6) {
			copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[1]
		} else if key.Matches(msg, m.keyMap.CopyDNSName) {
			copyStr = m.list.SelectedItem().(listItem).status.DNSName
		}
		if copyStr == "" {
			m.list.NewStatusMessage("Sorry, nothing to copy.")
		} else {
			clipboard.WriteAll(copyStr)
			m.list.NewStatusMessage(fmt.Sprintf("Copied \"%s\"!", constants.AccentTextStyle.Copy().Underline(true).Render(copyStr)))
		}
	}
	if key.Matches(msg, m.keyMap.Refresh) {
		cmd = func() tea.Msg { return RefreshMsg(true) }
		cmds = append(cmds, cmd)
	}
	if key.Matches(msg, m.keyMap.Enter) {
		cmd = func() tea.Msg { return NodeSelectedMsg(m.list.SelectedItem().(listItem).status.PublicKey) }
		cmds = append(cmds, cmd)
	}
	return m, cmds
}

func (m Model) headerView() string {
	hostname := ""
	userInfo := ""
	os := ""
	ips := ""
	offersExitNode := "no"
	usingExitNode := "-"
	if user, ok := m.tailStatus.User[strconv.FormatInt(m.tailStatus.Self.UserID, 10)]; ok {
		userInfo = fmt.Sprintf("%s <%s>", user.DisplayName, user.LoginName)
	}
	if e, ok := m.tailStatus.Peer[string(m.exitNode)]; ok {
		usingExitNode = constants.WarningTextStyle.Render(e.HostName)
	}
	hostname = m.tailStatus.Self.HostName
	os = m.tailStatus.Self.OS
	if m.tailStatus.Self.ExitNodeOption {
		offersExitNode = constants.WarningTextStyle.Render("yes")
	}
	ipList := m.tailStatus.TailscaleIPs
	ipList = append(ipList, m.tailStatus.Self.DNSName)
	ips = strings.Join(ipList, " | ")
	title := constants.TitleStyle.Render(" This node ")
	hostname = constants.AccentTextStyle.Render("Host: ") + hostname + " (" + os + ")"
	ips = constants.AccentTextStyle.Render("IPs: ") + ips
	exitNode := constants.AccentTextStyle.Render("Exit node: ") + constants.MutedTextStyle.Render("offers:") + (offersExitNode) + constants.MutedTextStyle.Render(" / using:") + usingExitNode
	body := constants.NormalTextStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s %s", userInfo, hostname, ips, exitNode, m.msg))
	return constants.HeaderStyle.Render(fmt.Sprintf("%s\n\n%s", title, body))
}

func (m *Model) getItems() []list.Item {
	items := []list.Item{}
	peers := []tailscale.PeerStatus{}
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

	peers = append([]tailscale.PeerStatus{m.tailStatus.Self}, peers...)

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
			owner = "from:" + m.tailStatus.User[strconv.FormatInt(v.UserID, 10)].LoginName
		}
		owner = constants.MutedTextStyle.Render("[" + owner + "]")
		exitNode := ""
		if v.ExitNodeOption {
			exitNode = constants.MutedTextStyle.Copy().Bold(true).Render("[→]")
		}
		if v.ExitNode {
			m.exitNode = v.PublicKey
			exitNode = constants.SuccessTextStyle.Copy().Bold(true).Render("[→]")
		}
		os := v.OS
		title := fmt.Sprintf("%s %s %s %s %s", hostName, state, owner, os, exitNode)
		desc := "- "
		ips := v.TailscaleIPs
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
	return fmt.Sprintf("%s\n%s", m.headerView(), m.list.View())
}

func New(status *tailscale.Status, w, h int) Model {
	d := list.NewDefaultDelegate()
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)
	d.Styles.NormalDesc = d.Styles.NormalTitle.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"}).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy()
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)
	d.Styles.DimmedDesc = d.Styles.DimmedTitle.Copy().
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
	d.SetHeight(2)
	d.SetSpacing(1)
	m := Model{
		list:       list.New([]list.Item{}, d, w, h),
		keyMap:     keymap.NewKeyMap(),
		tailStatus: status,
	}
	headerHeight := lipgloss.Height(m.headerView())
	m.list.SetHeight(h - headerHeight)

	m.list.SetItems(m.getItems())

	m.list.Title = "Nodes"
	m.list.Styles.Title = constants.TitleStyle.Copy().Padding(0, 1)
	m.list.FilterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
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
