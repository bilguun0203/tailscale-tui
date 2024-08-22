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
	tsKey "tailscale.com/types/key"
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
	tsLocalClient  *tailscale.LocalClient
	tsStatus       *ipnstate.Status
	selectedNodeID tsKey.NodePublic
	isLoading      bool
	Err            error
	nodelist       nodelist.Model
	nodedetails    nodedetails.Model
	spinner        spinner.Model
	w, h           int
}

type statusLoaded *ipnstate.Status
type statusError error

func (m Model) getTsStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.tsLocalClient.Status(context.Background())
		if err != nil {
			return statusError(err)
		}
		return statusLoaded(status)
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		m.getTsStatus(),
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tmpCmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case statusLoaded:
		m.isLoading = false
		m.Err = nil
		m.tsStatus = msg
		m.viewState = viewStateList
		m.nodelist = nodelist.New(m.tsStatus, m.w, m.h)
	case statusError:
		m.isLoading = false
		m.Err = msg
		return m, tea.Quit
	case nodedetails.BackMsg:
		m.viewState = viewStateList
	case nodelist.RefreshMsg:
		m.isLoading = true
		cmds = append(cmds, m.getTsStatus())
		cmds = append(cmds, m.spinner.Tick)
	case nodelist.NodeSelectedMsg:
		m.selectedNodeID = tsKey.NodePublic(msg)
		m.nodedetails = nodedetails.New(m.tsStatus, m.selectedNodeID, m.w, m.h)
		m.viewState = viewStateDetails
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if !m.isLoading {
			m.nodelist.SetSize(msg.Width, msg.Height)
		}
	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, tmpCmd = m.spinner.Update(msg)
			cmds = append(cmds, tmpCmd)
		}
	}

	switch m.viewState {
	case viewStateDetails:
		m.nodedetails, tmpCmd = m.nodedetails.Update(msg)
		cmds = append(cmds, tmpCmd)
	case viewStateList:
		if m.isLoading {
			m.spinner, tmpCmd = m.spinner.Update(msg)
			cmds = append(cmds, tmpCmd)
		} else {
			m.nodelist, tmpCmd = m.nodelist.Update(msg)
			cmds = append(cmds, tmpCmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.viewState {
	case viewStateDetails:
		return m.nodedetails.View()
	case viewStateList:
		if m.isLoading {
			return fmt.Sprintf("\n\n   %s Loading...\n\n", m.spinner.View())
		}
		return m.nodelist.View()
	default:
		return "*_*"
	}
}

func New(lc *tailscale.LocalClient) Model {
	m := Model{
		tsLocalClient: lc,
		viewState:     viewStateList,
		isLoading:     false,
		spinner:       spinner.New(),
	}
	m.spinner.Spinner = spinner.Dot
	m.spinner.Style = constants.SpinnerStyle

	m.nodelist = nodelist.New(nil, m.w, m.h)
	return m
}
