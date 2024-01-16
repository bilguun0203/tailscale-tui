package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
)

var layoutStyle = lipgloss.NewStyle().Margin(1)
var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
var headerStyle = lipgloss.NewStyle().Margin(2)

type listItem struct {
	title, desc string
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return i.desc }
func (i listItem) FilterValue() string { return i.title }

type model struct {
	loaded     bool
	err        error
	tailStatus *ipnstate.Status
	list       list.Model
	spinner    spinner.Model
}

type loadedStatus *ipnstate.Status

func getStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := tailscale.Status(context.Background())
		if err != nil {
			return err
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
	case loadedStatus:
		fmt.Println("status")
		m.tailStatus = msg
		m.loaded = true
		items := getItems(msg)
		m.list.NewStatusMessage(fmt.Sprintf("%d machines found!", len(items)))
		cmd = m.list.SetItems(items)
		cmds = append(cmds, cmd)
	case error:
		m.err = msg
		m.list.NewStatusMessage("Is Tailscale off?")
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
	if m.loaded {
		return layoutStyle.Render(fmt.Sprintf("%s\n%s", m.headerView(), m.list.View()))
	} else {
		str := fmt.Sprintf("\n\n   %s Loading ...\n\n", m.spinner.View())
		return layoutStyle.Render(str)
	}
}

func (m model) headerView() string {
	textStyle := lipgloss.NewStyle().Background(lipgloss.Color("205"))
	hostname := ""
	ips := ""
	if m.loaded && m.err == nil {
		hostname = m.tailStatus.Self.HostName
		for _, v := range m.tailStatus.TailscaleIPs {
			ips += "\n\t- " + v.String()
		}
	}

	return headerStyle.Render(fmt.Sprintf("%s\n\nHostname: %s\nIPs: %s", textStyle.Render(" My machine "), hostname, ips))
}

func getItems(status *ipnstate.Status) []list.Item {
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

	for _, v := range peers {
		dangerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e31743"))
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#47ed73"))
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999999"))
		state := dangerStyle.Render("●")
		if v.Online {
			state = successStyle.Render("●")
		}
		owner := "my device"
		if v.UserID != status.Self.UserID {
			owner = "from:" + status.User[v.UserID].LoginName
		}
		owner = mutedStyle.Render("[" + owner + "]")
		exitNode := ""
		if v.ExitNodeOption {
			exitNode = mutedStyle.Copy().Bold(true).Render("[→]")
		}
		if v.ExitNode {
			exitNode = successStyle.Copy().Bold(true).Render("[→]")
		}
		title := fmt.Sprintf("%s %s %s %s %s", state, v.HostName, owner, v.OS, exitNode)
		desc := ""
		for _, ip := range v.TailscaleIPs {
			desc += fmt.Sprintf("%s; ", ip)
		}
		items = append(items, listItem{title: title, desc: desc})
	}
	return items
}

func initialModel(ctx context.Context) model {
	m := model{list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0), spinner: spinner.New()}
	m.list.Title = "Machines"
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
