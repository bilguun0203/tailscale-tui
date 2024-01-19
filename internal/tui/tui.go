package tui

import (
	"context"
	"fmt"

	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	nodedetails "github.com/bilguun0203/tailscale-tui/internal/tui/node_details"
	nodelist "github.com/bilguun0203/tailscale-tui/internal/tui/node_list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"tailscale.com/client/tailscale"
	"tailscale.com/ipn/ipnstate"
	tskey "tailscale.com/types/key"
)

type viewState int

const (
	viewStateList viewState = iota
	viewStateDetails
)

func (f viewState) String() string {
	return [...]string{
		"list",
		"details",
	}[f]
}

type Model struct {
	viewState      viewState
	tsStatus       *ipnstate.Status
	selectedNodeID tskey.NodePublic
	isLoading      bool
	err            error
	nodelist       nodelist.Model
	nodedetails    nodedetails.Model
	spinner        spinner.Model
	msg            string
	w, h           int
}

type statusLoaded *ipnstate.Status
type statusError error
type nodeSelected *ipnstate.PeerStatus
type backNodeSelected bool

func getTsStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := tailscale.Status(context.Background())
		if err != nil {
			return statusError(err)
		}
		return statusLoaded(status)
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		getTsStatus(),
		m.spinner.Tick,
		m.nodelist.Init(),
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case statusLoaded:
		m.tsStatus = msg
		m.nodelist = nodelist.New(m.tsStatus, m.w, m.h)
		m.isLoading = false
	case statusError:
		m.isLoading = false
		return m, tea.Quit
	case nodedetails.BackMsg:
		m.viewState = viewStateList
	case nodelist.RefreshMsg:
		cmd = getTsStatus()
		cmds = append(cmds, cmd)
	case nodelist.NodeSelectedMsg:
		m.selectedNodeID = tskey.NodePublic(msg)
		m.nodedetails = nodedetails.New(m.tsStatus, m.selectedNodeID, m.w, m.h)
		m.viewState = viewStateDetails
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if !m.isLoading {
			m.nodelist.SetSize(msg.Width, msg.Height)
		}
	}

	switch m.viewState {
	case viewStateDetails:
		m.nodedetails, cmd = m.nodedetails.Update(msg)
	case viewStateList:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
		} else {
			m.nodelist, cmd = m.nodelist.Update(msg)
		}
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.viewState {
	case viewStateDetails:
		return m.nodedetails.View()
	default:
		if m.isLoading {
			return fmt.Sprintf("\n\n   %s Loading ...\n\n", m.spinner.View())
		}
		return m.nodelist.View()
	}
}

func New() Model {
	m := Model{
		viewState: viewStateList,
		isLoading: true,
		spinner:   spinner.New(),
	}
	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = constants.SpinnerStyle
	return m
}
