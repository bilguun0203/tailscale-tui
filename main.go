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
var mutedTextStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"})
var layoutStyle = lipgloss.NewStyle().Margin(1)
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var headerStyle = lipgloss.NewStyle().Margin(2)

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
	copyIpv4 key.Binding
}

func newListKeyMaps() *listKeyMap {
	return &listKeyMap{
		copyIpv4: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy ipv4"),
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
		if msg.String() == "y" {
			if len(m.list.SelectedItem().(listItem).status.TailscaleIPs) > 0 {
				copyStr := m.list.SelectedItem().(listItem).status.TailscaleIPs[0].String()
				clipboard.WriteAll(copyStr)
				m.list.NewStatusMessage(fmt.Sprintf("Copied \"%s\"!", copyStr))
			} else {
				m.list.NewStatusMessage("Sorry, nothing to copy.")
			}
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
	ips := ""
	offersExitNode := "No"
	exitNode := "-"
	if m.loaded && m.err == nil {
		if user, ok := m.tailStatus.User[m.tailStatus.Self.UserID]; ok {
			userInfo = fmt.Sprintf("%s <%s>", user.DisplayName, user.LoginName)
		}
		if e, ok := m.tailStatus.Peer[m.exitNode]; ok {
			exitNode = e.HostName
		}
		hostname = m.tailStatus.Self.HostName
		if m.tailStatus.Self.ExitNodeOption {
			offersExitNode = "Yes"
		}
		for _, v := range m.tailStatus.TailscaleIPs {
			ips += "\n- " + v.String()
		}
		ips += "\n- " + m.tailStatus.Self.DNSName
	}
	title := titleStyle.Render(" This node ")
	body := normalTextStyle.Render(fmt.Sprintf("%s\n\nHostname: %s\n%s\n\nOffers Exit Node: %s\nUsing Exit Node: %s", userInfo, hostname, ips, offersExitNode, exitNode))
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

	m.exitNode = tskey.NodePublic{}
	for _, v := range peers {
		state := dangerTextStyle.Render("●")
		if v.Online {
			state = successTextStyle.Render("●")
		}
		owner := "my device"
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
		title := fmt.Sprintf("%s %s %s %s %s", v.HostName, state, owner, v.OS, exitNode)
		desc := "- "
		for _, ip := range v.TailscaleIPs {
			desc += fmt.Sprintf("%s. ", ip)
		}
		desc += fmt.Sprintf("\n- %s", v.DNSName)
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
	d.SetHeight(3)

	m := model{list: list.New([]list.Item{}, d, 0, 0), spinner: spinner.New()}
	m.list.Title = "Nodes"
	m.list.Styles.Title = titleStyle.Copy().Padding(0, 1)
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMaps.copyIpv4,
		}
	}
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMaps.copyIpv4,
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
