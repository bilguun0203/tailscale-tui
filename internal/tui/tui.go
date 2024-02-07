package tui

import (
	"fmt"

	"github.com/bilguun0203/tailscale-tui/internal/tailscale"
	"github.com/bilguun0203/tailscale-tui/internal/tui/constants"
	nodedetails "github.com/bilguun0203/tailscale-tui/internal/tui/node_details"
	nodelist "github.com/bilguun0203/tailscale-tui/internal/tui/node_list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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
	tsStatus       tailscale.Status
	selectedNodeID string
	isLoading      bool
	err            error
	nodelist       nodelist.Model
	nodedetails    nodedetails.Model
	spinner        spinner.Model
	w, h           int
}

type statusRequest struct {
	status tailscale.Status
	err    error
}

type statusLoaded tailscale.Status
type statusError error

func getTsStatusAsync(c chan statusRequest) {
	ts, err := tailscale.New()
	if err != nil {
		c <- statusRequest{status: tailscale.Status{}, err: err}
		return
	}
	s, e := ts.Status()
	c <- statusRequest{status: s, err: e}
}

func getTsStatus() tea.Cmd {
	return func() tea.Msg {
		c := make(chan statusRequest)
		go getTsStatusAsync(c)
		statusReq := <-c
		if statusReq.err != nil {
			return statusError(statusReq.err)
		}
		return statusLoaded(statusReq.status)
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
		m.tsStatus = tailscale.Status(msg)
		m.nodelist = nodelist.New(&m.tsStatus, m.w, m.h)
		m.isLoading = false
	case statusError:
		m.isLoading = false
		m.err = msg
		return m, tea.Quit
	case nodedetails.BackMsg:
		m.viewState = viewStateList
	case nodelist.RefreshMsg:
		cmd = getTsStatus()
		cmds = append(cmds, cmd)
	case nodelist.NodeSelectedMsg:
		m.selectedNodeID = string(msg)
		m.nodedetails = nodedetails.New(&m.tsStatus, m.selectedNodeID, m.w, m.h)
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
