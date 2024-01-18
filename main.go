package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
	tskey "tailscale.com/types/key"
)

var titleStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#20F394")).Foreground(lipgloss.Color("#000000"))
var normalTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"})
var dangerTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "197", Dark: "197"})
var successTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "034", Dark: "049"})
var warningTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "214", Dark: "214"})
var mutedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
var accentTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
var layoutStyle = lipgloss.NewStyle()
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var headerStyle = lipgloss.NewStyle().Margin(1, 2)

type listItem struct {
	title, desc string
	status      ipnstate.PeerStatus
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title + " " + i.desc }

type model struct {
	loaded     bool
	err        error
	tailStatus *ipnstate.Status
	exitNode   tskey.NodePublic
	list       list.Model
	spinner    spinner.Model
}

type loadedStatus *ipnstate.Status

type statusError error

type listKeyMap struct {
	copyIpv4    key.Binding
	copyIpv6    key.Binding
	copyDNSName key.Binding
	refresh     key.Binding
}

func newListKeyMaps() *listKeyMap {
	return &listKeyMap{
		copyIpv4: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy ipv4"),
		),
		copyIpv6: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "copy ipv6"),
		),
		copyDNSName: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy domain"),
		),
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

func getStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := tailscale.Status(context.Background())
		if err != nil {
			return statusError(err)
		}
		return loadedStatus(status)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(getStatus(), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "y" || msg.String() == "Y" || msg.String() == "ctrl+y" {
			copyStr := ""
			ipCount := len(m.list.SelectedItem().(listItem).status.TailscaleIPs)
			if ipCount > 0 && msg.String() == "y" {
				copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[0].String()
			} else if ipCount > 1 && msg.String() == "Y" {
				copyStr = m.list.SelectedItem().(listItem).status.TailscaleIPs[1].String()
			} else if msg.String() == "ctrl+y" {
				copyStr = m.list.SelectedItem().(listItem).status.DNSName
			}
			if copyStr == "" {
				m.list.NewStatusMessage("Sorry, nothing to copy.")
			} else {
				clipboard.WriteAll(copyStr)
				m.list.NewStatusMessage(fmt.Sprintf("Copied \"%s\"!", accentTextStyle.Copy().Underline(true).Render(copyStr)))
			}
		}
		if msg.String() == "r" {
			cmd = getStatus()
			cmds = append(cmds, cmd)
		}
	case loadedStatus:
		m.tailStatus = msg
		m.loaded = true
		items := getItems(&m, msg)
		cmd = m.list.SetItems(items)
		cmds = append(cmds, cmd)
	case statusError:
		m.err = msg
		m.loaded = true
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		h, v := layoutStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-headerHeight)
	default:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		str := fmt.Sprintf("Failed to get Tailscale status.\nIs your Tailscale active?\n\n%s", mutedTextStyle.Render("press q to quit"))
		return layoutStyle.Render(str)
	}
	if m.loaded {
		return layoutStyle.Render(fmt.Sprintf("%s\n%s", m.headerView(), m.list.View()))
	}
	str := fmt.Sprintf("\n\n   %s Loading ...\n\n", m.spinner.View())
	return layoutStyle.Render(str)
}

func (m model) headerView() string {
	hostname := ""
	userInfo := ""
	os := ""
	ips := ""
	offersExitNode := "no"
	usingExitNode := "-"
	if m.loaded && m.err == nil {
		if user, ok := m.tailStatus.User[m.tailStatus.Self.UserID]; ok {
			userInfo = fmt.Sprintf("%s <%s>", user.DisplayName, user.LoginName)
		}
		if e, ok := m.tailStatus.Peer[m.exitNode]; ok {
			usingExitNode = warningTextStyle.Render(e.HostName)
		}
		hostname = m.tailStatus.Self.HostName
		os = m.tailStatus.Self.OS
		if m.tailStatus.Self.ExitNodeOption {
			offersExitNode = warningTextStyle.Render("yes")
		}
		ipList := []string{}
		for _, ip := range m.tailStatus.TailscaleIPs {
			ipList = append(ipList, ip.String())
		}
		ipList = append(ipList, m.tailStatus.Self.DNSName)
		ips = strings.Join(ipList, " | ")
	}
	title := titleStyle.Render(" This node ")
	hostname = accentTextStyle.Render("Host: ") + hostname + " " + os
	ips = accentTextStyle.Render("IPs: ") + ips
	exitNode := accentTextStyle.Render("Exit node: ") + mutedTextStyle.Render("offers:") + (offersExitNode) + mutedTextStyle.Render(" / using:") + usingExitNode
	body := normalTextStyle.Render(fmt.Sprintf("%s\n\n%s\n%s\n%s", userInfo, hostname, ips, exitNode))
	return headerStyle.Render(fmt.Sprintf("%s\n\n%s", title, body))
}

func getItems(m *model, status *ipnstate.Status) []list.Item {
	items := []list.Item{}
	peers := []ipnstate.PeerStatus{}
	for _, v := range status.Peer {
		peers = append(peers, *v)
	}
	sort.Slice(peers, func(i, j int) bool {
		ownNode1 := status.Self.UserID == peers[i].UserID
		ownNode2 := status.Self.UserID == peers[j].UserID
		first := fmt.Sprint(!ownNode1) + strings.ToLower(peers[i].HostName)
		second := fmt.Sprint(!ownNode2) + strings.ToLower(peers[j].HostName)
		return first < second
	})

	peers = append([]ipnstate.PeerStatus{*status.Self}, peers...)

	m.exitNode = tskey.NodePublic{}
	for _, v := range peers {
		state := dangerTextStyle.Render("●")
		if v.Online {
			state = successTextStyle.Render("●")
		}
		hostName := v.HostName
		owner := "my device"
		if v.ID == status.Self.ID {
			hostName = "◆ " + hostName
			owner = "this device"
		}
		if v.UserID != status.Self.UserID {
			owner = "from:" + status.User[v.UserID].LoginName
		}
		owner = mutedTextStyle.Render("[" + owner + "]")
		exitNode := ""
		if v.ExitNodeOption {
			exitNode = mutedTextStyle.Copy().Bold(true).Render("[→]")
		}
		if v.ExitNode {
			m.exitNode = v.PublicKey
			exitNode = successTextStyle.Copy().Bold(true).Render("[→]")
		}
		os := v.OS
		title := fmt.Sprintf("%s %s %s %s %s", hostName, state, owner, os, exitNode)
		desc := "- "
		ips := []string{}
		for _, ip := range v.TailscaleIPs {
			ips = append(ips, ip.String())
		}
		ips = append(ips, v.DNSName)
		desc += strings.Join(ips, " | ")
		items = append(items, listItem{title: title, desc: desc, status: v})
	}
	return items
}

func initialModel(ctx context.Context) model {
	keyMaps := newListKeyMaps()
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

	m := model{list: list.New([]list.Item{}, d, 0, 0), spinner: spinner.New()}
	m.list.Title = "Nodes"
	m.list.Styles.Title = titleStyle.Copy().Padding(0, 1)
	m.list.FilterInput.PromptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
	m.list.FilterInput.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#00C86E", Dark: "#20F394"})
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMaps.copyIpv4,
			keyMaps.refresh,
		}
	}
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMaps.copyIpv4,
			keyMaps.copyIpv6,
			keyMaps.copyDNSName,
			keyMaps.refresh,
		}
	}
	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = spinnerStyle
	return m
}

func main() {
	ctx := context.Background()
	m := initialModel(ctx)

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
